package statistic

import (
	"time"

	"github.com/jansorg/marketplace-stats/marketplace"
)

type Month struct {
	Date      time.Time
	PrevMonth *Month

	DownloadsTotal  int
	DownloadsUnique int
	TotalSalesUSD   AmountAndFee
	NewSalesUSD     AmountAndFee

	// Customers who bought in the current month
	Customers        int
	CustomersAnnual  int
	CustomersMonthly int
	CustomersMap     marketplace.CustomersMap

	// New customers, who never bought before
	NewCustomers        int
	NewCustomersAnnual  int
	NewCustomersMonthly int

	HasChurnRate         bool
	ChurnActiveCustomers int
	ChurnedCustomers     marketplace.ChurnedCustomers

	ActiveCustomersMonthly int
	ActiveCustomersAnnual  int
	ActiveCustomersTotal   int

	AnnualRevenueUSD AmountAndFee
}

func NewMonth(year int, month time.Month) *Month {
	return &Month{
		Date: time.Date(year, month, 1, 0, 0, 0, 0, marketplace.ServerTimeZone),
	}
}

func NewMonthForDate(date time.Time) *Month {
	return NewMonth(date.Year(), date.Month())
}

func (m *Month) Name() string {
	return m.Date.Format("2006-01")
}

func (m *Month) IsActiveMonth() bool {
	now := time.Now().In(marketplace.ServerTimeZone)
	return m.Date.Year() == now.Year() && m.Date.Month() == now.Month()
}

func (m *Month) PreviousYearMonth() *Month {
	month := m
	i := 12
	for i > 1 {
		month = month.PrevMonth
		if month == nil {
			return nil
		}
		i--
	}
	return month
}

func (m *Month) DownloadsTotalRatio(downloads int) marketplace.Amount {
	return m.TotalSalesUSD.Total / marketplace.Amount(downloads)
}

// IsChurned returns if the given customer churned in this month or in a previous month
// If it churned before, but bought again later, then it's considered as not churned.
// fixme this is potentially slow
func (m *Month) IsChurned(id marketplace.CustomerID) bool {
	if _, active := m.CustomersMap[id]; active {
		return false
	}

	if m.ChurnedCustomers.Contains(id) {
		return true
	}

	if m.PrevMonth != nil {
		return m.PrevMonth.IsChurned(id)
	}

	return false
}

func (m *Month) ChurnRatePercentage() float64 {
	return float64(len(m.ChurnedCustomers)) / float64(m.ChurnActiveCustomers) * 100
}

func findMonth(downloads []marketplace.DownloadMonthly, yearMonth time.Time) int {
	y, m, _ := yearMonth.Date()
	for _, d := range downloads {
		if d.Year == y && d.Month == m {
			return d.Downloads
		}
	}
	return 0
}

// Update the current month's data from the complete collection of sales
func (m *Month) Update(sales marketplace.Sales, previousMonthData *Month, downloadsTotal []marketplace.DownloadMonthly, downloadsUnique []marketplace.DownloadMonthly) {
	m.PrevMonth = previousMonthData

	// find download counts
	m.DownloadsTotal = findMonth(downloadsTotal, m.Date)
	m.DownloadsUnique = findMonth(downloadsUnique, m.Date)

	allPreviousSales := sales.Before(m.Date)
	currentMonthSales := sales.ByMonth(m.Date.Year(), m.Date.Month())

	// Basics
	m.CustomersMap = currentMonthSales.CustomersMap()
	m.Customers = len(m.CustomersMap)
	m.CustomersAnnual = len(currentMonthSales.ByAnnualSubscription().CustomersMap())
	m.CustomersMonthly = len(currentMonthSales.ByMonthlySubscription().CustomersMap())

	// New customers
	newCustomerSales := currentMonthSales.ByNewCustomers(allPreviousSales, m.Date)
	newAnnualCustomersMap := newCustomerSales.ByAnnualSubscription().CustomersMap()
	newMonthlyCustomersMap := newCustomerSales.ByMonthlySubscription().CustomersMap()
	m.NewCustomersAnnual = len(newAnnualCustomersMap)
	m.NewCustomersMonthly = len(newMonthlyCustomersMap.Without(newAnnualCustomersMap))
	m.NewCustomers = m.NewCustomersMonthly + m.NewCustomersAnnual

	// Sales
	m.TotalSalesUSD.Total = currentMonthSales.TotalSumUSD()
	m.TotalSalesUSD.Fee = currentMonthSales.FeeSumUSD()

	m.NewSalesUSD.Total = newCustomerSales.TotalSumUSD()
	m.NewSalesUSD.Fee = newCustomerSales.FeeSumUSD()

	// churn, no churn for first month
	m.HasChurnRate = m.PrevMonth != nil
	if m.HasChurnRate {
		// JetBrains mentioned 7 days as grace time for expired licenses
		m.ChurnActiveCustomers, m.ChurnedCustomers = computeMonthlyChurn(marketplace.NewYearMonthByDate(m.Date), sales, 7)
	}

	// Active customers
	if previousMonthData != nil {
		m.ActiveCustomersMonthly = previousMonthData.ActiveCustomersMonthly - m.ChurnedCustomers.CountMonthly() + m.NewCustomersMonthly
		m.ActiveCustomersAnnual = previousMonthData.ActiveCustomersAnnual - m.ChurnedCustomers.CountAnnual() + m.NewCustomersAnnual
		m.ActiveCustomersTotal = previousMonthData.ActiveCustomersTotal - len(m.ChurnedCustomers) + m.NewCustomers
	} else {
		m.ActiveCustomersMonthly = m.NewCustomersMonthly
		m.ActiveCustomersAnnual = m.NewCustomersAnnual
		m.ActiveCustomersTotal = m.NewCustomers
	}

	// projected annual revenue (ARR)
	prevYearSales := sales.ByDateRange(
		marketplace.NewYearMonthDayByDate(m.Date.AddDate(-1, 0, 0)),
		marketplace.NewYearMonthDayByDate(m.Date))
	for _, sale := range prevYearSales.ByAnnualSubscription() {
		m.AnnualRevenueUSD.Total += sale.AmountUSD
	}

	// estimate if the current month isn't finished yet
	monthSalesFactor := 1.0
	if m.IsActiveMonth() {
		y, m, d := time.Now().In(marketplace.ServerTimeZone).Date()
		lastDay := time.Date(y, m, 1, 0, 0, 0, 0, marketplace.ServerTimeZone).AddDate(0, 1, -1).Day()
		monthSalesFactor = float64(lastDay) / float64(d)
	}

	for _, sale := range currentMonthSales {
		if sale.Period == marketplace.MonthlySubscription {
			m.AnnualRevenueUSD.Total += sale.AmountUSD * 12 * marketplace.Amount(monthSalesFactor)
		} else if sale.Period == marketplace.AnnualSubscription {
			m.AnnualRevenueUSD.Total += sale.AmountUSD * marketplace.Amount(monthSalesFactor)
		}
	}

	// try to estimate for missing months of the year
	//monthsWithSales := m.Date.In(marketplace.ServerTimeZone).Month() - currentYearSales[0].Date.Month() + 1
	//if monthsWithSales >= 1 && monthsWithSales <= 11 {
	//	m.AnnualRevenueUSD.Total *= marketplace.Amount(1.0 + float64(12-monthsWithSales)/12.0)
	//}

	m.AnnualRevenueUSD.Total *= 0.8 // 20% discount in 2nd and 3rd year, fixme handle 4th+ year
	m.AnnualRevenueUSD.Fee = m.AnnualRevenueUSD.Total * marketplace.Amount(marketplace.FeePercentage(m.Date.AddDate(1, 0, 0)))
}

func computeMonthlyChurn(month marketplace.YearMonth, allSales marketplace.Sales, graceDays int) (int, marketplace.ChurnedCustomers) {
	if len(allSales) == 0 {
		return 0.0, nil
	}

	// allow 3 days into the following month as grace time
	now := time.Now().In(marketplace.ServerTimeZone)
	isCurrentMonth := month.ContainsDate(now)
	var totalUsers int
	var churned marketplace.ChurnedCustomers

	// all sales of current month + grace time
	graceTimeEnd := month.AsDate().AddDate(0, 1, graceDays)
	salesNew := allSales.AtOrAfter(month.AsDate()).Before(graceTimeEnd).CustomerSalesMap()

	// all customers of the previous month, who didn't renew the current month
	previousMonth := month.PreviousMonth()
	previousMonthCustomers := allSales.ByYearMonth(previousMonth).ByMonthlySubscription().Customers()
	if len(previousMonthCustomers) > 0 {
		totalUsers += len(previousMonthCustomers)
		previousMonthSales := allSales.ByYearMonth(previousMonth).ByMonthlySubscription().CustomerSalesMap()

		upgradeSales := allSales.AtOrAfter(previousMonth.AsDate()).Before(month.AsDate()).ByAnnualSubscription().CustomersMap()
		for _, candidate := range previousMonthCustomers {
			_, upgraded := upgradeSales[candidate.ID]
			_, boughtAgain := salesNew[candidate.ID]
			if !upgraded && !boughtAgain {
				// in the current month check grace time before recording as churned
				if !isCurrentMonth || previousMonthSales[candidate.ID].LatestPurchase().AddDate(0, 1, graceDays).Before(now) {
					churned = append(churned, marketplace.ChurnedCustomer{
						Customer:     candidate,
						LastPurchase: marketplace.NewYearMonthDayByDate(previousMonthSales[candidate.ID].LatestPurchase()),
						Subscription: marketplace.MonthlySubscription,
					})
				}
			}
		}
	}

	// all customers with an annual subscription, who didn't renew in the current month
	previousYearMonth := previousMonth.Add(-1, 0, 0)
	lastPurchases := allSales.Before(month.AsDate()).CustomersLastPurchase()
	// only keep sales of previous year's month, if no later purchases were made by the same customer between then and the current month
	// this is only an estimate, because it's not possible to track subscriptions, only customers
	expectedAnnual := allSales.ByAnnualSubscription().ByYearMonth(previousYearMonth).FilterBy(func(sale marketplace.Sale) bool {
		lastPurchase, found := lastPurchases[sale.Customer.ID]
		return !found || !lastPurchase.IsAfter(sale.Date)
	})
	if len(expectedAnnual) > 0 {
		previousYearCustomers := expectedAnnual.Customers()
		prevYearSales := expectedAnnual.CustomerSalesMap()
		totalUsers += len(previousYearCustomers)
		for _, candidate := range previousYearCustomers {
			_, boughtAgain := salesNew[candidate.ID]
			if !boughtAgain {
				if !isCurrentMonth || prevYearSales[candidate.ID].LatestPurchase().AddDate(1, 0, graceDays).Before(now) {
					churned = append(churned, marketplace.ChurnedCustomer{
						Customer:     candidate,
						LastPurchase: lastPurchases[candidate.ID],
						Subscription: marketplace.AnnualSubscription,
					})
				}
			}
		}
	}

	return totalUsers, churned
}
