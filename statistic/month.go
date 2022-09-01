package statistic

import (
	"time"

	"github.com/jansorg/marketplace-stats/marketplace"
)

type Month struct {
	Date          time.Time
	PreviousMonth *Month

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

	// Trial conversions, these are totals from the beginning to the end of the current month
	TrialCountMonth       int
	TrialCountTotal       int
	TrialConversionsTotal int

	// Churned Customers
	ChurnedAnnual  marketplace.ChurnedCustomers
	ChurnedMonthly marketplace.ChurnedCustomers

	// Returned customers, who churned before and are back this month
	ReturnedCustomers        int
	ReturnedCustomersAnnual  int
	ReturnedCustomersMonthly int
	ReturnedCustomersList    marketplace.Customers

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
	i := 0
	for i < 12 {
		month = month.PreviousMonth
		if month == nil {
			return nil
		}
		i++
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

	if m.ChurnedAnnual.Contains(id) || m.ChurnedMonthly.Contains(id) {
		return true
	}

	if m.PreviousMonth != nil {
		return m.PreviousMonth.IsChurned(id)
	}

	return false
}

func (m *Month) AllChurned() marketplace.ChurnedCustomers {
	return marketplace.ChurnedCustomers{
		ChurnedCustomers: append(m.ChurnedMonthly.ChurnedCustomers, m.ChurnedAnnual.ChurnedCustomers...),
		ActiveUserCount:  m.ChurnedMonthly.ActiveUserCount + m.ChurnedAnnual.ActiveUserCount,
	}
}

// AllAnnualChurnedYear returns all users who churned in the previous 12 months, not just newly churned in the current month
// This doesn't contain free licenses, which were not renewed.
func (m *Month) AllAnnualChurnedYear() marketplace.ChurnedCustomers {
	seen := make(map[marketplace.CustomerID]bool)
	var churned []marketplace.ChurnedCustomer

	i := 0
	month := m
	for i < 12 && month != nil {
		for _, c := range month.ChurnedAnnual.ChurnedCustomers {
			if !seen[c.ID] && !c.FreeSubscription {
				churned = append(churned, c)
				seen[c.ID] = true
			}
		}
		i++
		month = month.PreviousMonth
	}

	return marketplace.ChurnedCustomers{
		ChurnedCustomers: churned,
		ActiveUserCount:  m.ChurnedAnnual.ActiveUserCount,
	}
}

// CollectAllChurned returns all churned users of previous periods, which did not return until the current period
func (m *Month) CollectAllChurned() marketplace.ChurnedCustomerList {
	if m.PreviousMonth == nil {
		return marketplace.ChurnedCustomerList{}
	}

	currentCustomers := m.CustomersMap
	churned := make(map[marketplace.CustomerID]marketplace.ChurnedCustomer)

	if m.PreviousMonth != nil {
		prevMonthChurned := m.PreviousMonth.CollectAllChurned()
		for _, c := range prevMonthChurned {
			_, seen := currentCustomers[c.ID]
			if !seen {
				churned[c.ID] = c
			}
		}

		for _, c := range m.AllChurned().ChurnedCustomers {
			_, seen := currentCustomers[c.ID]
			if !seen {
				churned[c.ID] = c
			}
		}
	}

	var churnedList marketplace.ChurnedCustomerList
	for _, c := range churned {
		churnedList = append(churnedList, c)
	}
	return churnedList
}

func (m *Month) AllReturnedCustomers() marketplace.Customers {
	result := m.ReturnedCustomersList
	result.SortByID()
	return result
}

func (m *Month) HasAnnualChurnRate() bool {
	return m.PreviousYearMonth() != nil
}

func (m *Month) HasMonthlyChurnRate() bool {
	return m.PreviousMonth != nil
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
func (m *Month) Update(sales marketplace.Sales, trials marketplace.Transactions, previousMonthData *Month, downloadsTotal []marketplace.DownloadMonthly, downloadsUnique []marketplace.DownloadMonthly, graceDays int) {
	m.PreviousMonth = previousMonthData

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

	// Returned customers
	if previousMonthData != nil {
		allChurned := previousMonthData.CollectAllChurned()
		returnedCustomerSales := currentMonthSales.ByReturnedCustomers(allChurned)
		returnedAnnualCustomers := returnedCustomerSales.ByAnnualSubscription().CustomersMap()
		returnedMonthlyCustomers := returnedCustomerSales.ByMonthlySubscription().CustomersMap()

		m.ReturnedCustomersAnnual = len(returnedAnnualCustomers)
		m.ReturnedCustomersMonthly = len(returnedMonthlyCustomers.Without(returnedAnnualCustomers))
		m.ReturnedCustomers = m.ReturnedCustomersMonthly + m.ReturnedCustomersAnnual
		m.ReturnedCustomersList = returnedCustomerSales.Customers()
	}

	// Sales
	m.TotalSalesUSD.Total = currentMonthSales.TotalSumUSD()
	m.TotalSalesUSD.Fee = currentMonthSales.FeeSumUSD()

	m.NewSalesUSD.Total = newCustomerSales.TotalSumUSD()
	m.NewSalesUSD.Fee = newCustomerSales.FeeSumUSD()

	// Trials, until end of month
	nextMonth := m.Date.AddDate(0, 1, 0)
	matchingTrials := trials.Before(nextMonth)
	monthTrials := matchingTrials.ByYearMonth(marketplace.NewYearMonth(m.Date.Year(), m.Date.Month()))
	convertedTrials := matchingTrials.GroupByCustomer().RetainCustomers(sales.Before(nextMonth).Customers())
	m.TrialCountMonth = len(monthTrials)
	m.TrialCountTotal = len(matchingTrials)
	m.TrialConversionsTotal = len(convertedTrials)

	// Churn, JetBrains said that there's a 7 days grace time for expired licenses
	if m.HasAnnualChurnRate() {
		m.ChurnedAnnual = computeAnnualChurn(marketplace.NewYearMonthByDate(m.Date), sales, graceDays)
		m.ChurnedAnnual.ActiveUserCount = m.PreviousYearMonth().ActiveCustomersAnnual
	}
	if m.HasMonthlyChurnRate() {
		m.ChurnedMonthly = computeMonthlyChurn(marketplace.NewYearMonthByDate(m.Date), sales, graceDays)
		m.ChurnedMonthly.ActiveUserCount = m.PreviousMonth.ActiveCustomersMonthly
	}

	// Active customers
	m.ActiveCustomersMonthly = m.NewCustomersMonthly + m.ReturnedCustomersMonthly
	m.ActiveCustomersAnnual = m.NewCustomersAnnual + m.ReturnedCustomersAnnual
	m.ActiveCustomersTotal = m.NewCustomers + m.ReturnedCustomers
	if previousMonthData != nil {
		m.ActiveCustomersMonthly += previousMonthData.ActiveCustomersMonthly - m.ChurnedMonthly.Count()
		m.ActiveCustomersAnnual += previousMonthData.ActiveCustomersAnnual - m.ChurnedAnnual.Count()
		m.ActiveCustomersTotal += previousMonthData.ActiveCustomersTotal - m.ChurnedAnnual.Count() - m.ChurnedMonthly.Count()
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

// returns number of active users and the churned users
func computeMonthlyChurn(month marketplace.YearMonth, allSales marketplace.Sales, graceDays int) marketplace.ChurnedCustomers {
	if len(allSales) == 0 {
		return marketplace.ChurnedCustomers{}
	}

	now := time.Now().In(marketplace.ServerTimeZone)
	var churned []marketplace.ChurnedCustomer

	// all sales of current month + grace time
	graceTimeEnd := month.AsDate().AddDate(0, 1, graceDays)
	salesGraceTime := allSales.AtOrAfter(month.AsDate()).Before(graceTimeEnd).CustomerSalesMap()
	previousMonth := month.PreviousMonth()

	// all customers of the previous month, who didn't renew the current month
	previousMonthCustomers := allSales.ByYearMonth(previousMonth).ByMonthlySubscription().Customers()
	if len(previousMonthCustomers) > 0 {
		previousMonthSales := allSales.ByYearMonth(previousMonth).ByMonthlySubscription().CustomerSalesMap()

		upgradeSales := allSales.AtOrAfter(previousMonth.AsDate()).Before(month.AsDate()).ByAnnualSubscription().CustomersMap()
		for _, candidate := range previousMonthCustomers {
			_, upgraded := upgradeSales[candidate.ID]
			_, boughtAgain := salesGraceTime[candidate.ID]
			saleExpected := previousMonthSales[candidate.ID].LatestPurchase().AddDate(0, 1, graceDays).Before(now)
			if !upgraded && !boughtAgain && saleExpected {
				churned = append(churned, marketplace.ChurnedCustomer{
					Customer:         candidate,
					LastPurchase:     marketplace.NewYearMonthDayByDate(previousMonthSales[candidate.ID].LatestPurchase()),
					Subscription:     marketplace.MonthlySubscription,
					FreeSubscription: previousMonthSales[candidate.ID].TotalUSD.IsZero(),
				})
			}
		}
	}
	return marketplace.NewChurnedCustomers(churned)
}

// returns users with annual subscriptions, who churned in the given month
func computeAnnualChurn(month marketplace.YearMonth, allSales marketplace.Sales, graceDays int) marketplace.ChurnedCustomers {
	now := time.Now().In(marketplace.ServerTimeZone)

	// all sales of current month + grace time
	graceTimeEnd := month.AsDate().AddDate(0, 1, graceDays)
	salesGraceTime := allSales.AtOrAfter(month.AsDate()).Before(graceTimeEnd).CustomerSalesMap()

	// all customers with an annual subscription, who didn't renew in the current month
	monthPreviousYear := month.Add(-1, 0, 0)
	lastPurchases := allSales.AtOrAfter(monthPreviousYear.AsDate()).Before(month.AsDate()).CustomersLastPurchase()
	// Only keep sales of previous year's month if no later purchases were made by the same customer between then and the current month.
	// This is only an estimate, because it's not possible to track subscriptions, only customers
	expectedAnnual := allSales.ByYearMonth(monthPreviousYear).ByAnnualSubscription().FilterBy(func(sale marketplace.Sale) bool {
		lastPurchase, found := lastPurchases[sale.Customer.ID]
		return !found || !lastPurchase.IsAfter(sale.Date)
	})

	previousYearCustomers := expectedAnnual.Customers()
	prevYearSales := expectedAnnual.CustomerSalesMap()

	var churned []marketplace.ChurnedCustomer
	for _, candidate := range previousYearCustomers {
		expected := prevYearSales[candidate.ID].LatestPurchase().AddDate(1, 0, graceDays).Before(now)
		_, boughtAgain := salesGraceTime[candidate.ID]
		if expected && !boughtAgain {
			churned = append(churned, marketplace.ChurnedCustomer{
				Customer:         candidate,
				LastPurchase:     lastPurchases[candidate.ID],
				Subscription:     marketplace.AnnualSubscription,
				FreeSubscription: prevYearSales[candidate.ID].TotalUSD.IsZero(),
			})
		}
	}
	return marketplace.NewChurnedCustomers(churned)
}
