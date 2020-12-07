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

	// Customers who bought in the current month
	Customers        int
	CustomersAnnual  int
	CustomersMonthly int

	// New customers, who never bought before
	NewCustomers        int
	NewCustomersAnnual  int
	NewCustomersMonthly int

	ChurnRate           float64
	ChurnRatePercentage float64
	ChurnedCustomers    marketplace.Customers
	HasChurnRate        bool

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

func (m *Month) ChurnRateTotalCustomers() int {
	return m.PrevMonth.ActiveCustomersMonthly
}

func (m *Month) DownloadsTotalRatio(downloads int) marketplace.Amount {
	return m.TotalSalesUSD.Total / marketplace.Amount(downloads)
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
	currentYearSales := sales.ByYear(m.Date.Year())

	// Basics
	m.Customers = len(currentMonthSales.CustomersMap())
	m.CustomersAnnual = len(currentMonthSales.ByAnnualSubscription().CustomersMap())
	m.CustomersMonthly = len(currentMonthSales.ByMonthlySubscription().CustomersMap())

	// New customers
	newAnnualCustomersMap := currentMonthSales.ByNewCustomers(allPreviousSales, m.Date).ByAnnualSubscription().CustomersMap()
	m.NewCustomersAnnual = len(newAnnualCustomersMap)
	m.NewCustomersMonthly = len(currentMonthSales.ByNewCustomers(allPreviousSales, m.Date).ByMonthlySubscription().CustomersMap().Without(newAnnualCustomersMap))
	m.NewCustomers = m.NewCustomersMonthly + m.NewCustomersAnnual

	// Sales
	m.TotalSalesUSD.Total = currentMonthSales.TotalSumUSD()
	m.TotalSalesUSD.Fee = currentMonthSales.FeeSumUSD()

	// churn, no churn for first month
	m.HasChurnRate = m.PrevMonth != nil
	if m.HasChurnRate {
		m.ChurnRate, m.ChurnedCustomers = computeMonthlyChurn(marketplace.NewYearMonthByDate(m.Date, marketplace.ServerTimeZone), sales, 3)
		m.ChurnRatePercentage = m.ChurnRate * 100
	}

	// Active customers
	if previousMonthData != nil {
		m.ActiveCustomersMonthly = previousMonthData.ActiveCustomersMonthly - len(m.ChurnedCustomers) + m.NewCustomersMonthly
		m.ActiveCustomersAnnual = previousMonthData.ActiveCustomersAnnual + m.NewCustomersAnnual
		m.ActiveCustomersTotal = previousMonthData.ActiveCustomersTotal - len(m.ChurnedCustomers) + m.NewCustomers
	} else {
		m.ActiveCustomersMonthly = m.NewCustomersMonthly
		m.ActiveCustomersAnnual = m.NewCustomersAnnual
		m.ActiveCustomersTotal = m.NewCustomers
	}

	// projected annual revenue
	if len(currentYearSales) > 0 {
		for _, sale := range currentYearSales.Before(m.Date).ByAnnualSubscription() {
			m.AnnualRevenueUSD.Total += sale.AmountUSD
		}

		// factor, if the current month isn't finished yet
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
}

func computeMonthlyChurn(month marketplace.YearMonth, allSales marketplace.Sales, graceDays int) (float64, []marketplace.Customer) {
	if len(allSales) == 0 {
		return 0.0, nil
	}

	// all customers of the previous month, who didn't upgrade and didn't buy in the current month
	previousMonth := month.PreviousMonth()
	previousMonthCustomers := allSales.ByYearMonth(previousMonth).ByMonthlySubscription().Customers()
	previousMonthSales := allSales.ByYearMonth(previousMonth).ByMonthlySubscription().CustomerSalesMap()

	// allow 3 days into the following month as grace time
	graceTimeEnd := month.AsDate().AddDate(0, 1, graceDays)
	upgradeSales := allSales.AtOrAfter(previousMonth.AsDate()).Before(month.AsDate()).ByAnnualSubscription().CustomersMap()
	salesNew := allSales.AtOrAfter(month.AsDate()).Before(graceTimeEnd).CustomerSalesMap()

	now := time.Now().In(marketplace.ServerTimeZone)
	isCurrentMonth := month.ContainsDate(now)

	var churned []marketplace.Customer
	for _, candidate := range previousMonthCustomers {
		_, upgraded := upgradeSales[candidate.ID]
		_, boughtAgain := salesNew[candidate.ID]
		if !upgraded && !boughtAgain {
			if !isCurrentMonth || previousMonthSales[candidate.ID].LatestPurchase().AddDate(0, 1, graceDays).Before(now) {
				churned = append(churned, candidate)
			}
		}
	}

	totalUsers := len(previousMonthCustomers)
	return float64(len(churned)) / float64(totalUsers), churned
}
