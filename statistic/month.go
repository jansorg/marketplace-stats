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

	// Customers who bought in the month
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

	AnnualRevenueUSD marketplace.Amount
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
	return m.CustomersMonthly + len(m.ChurnedCustomers)
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

	previousMonth := m.Date.AddDate(0, -1, 0)

	allPreviousSales := sales.Before(m.Date)
	previousMonthSales := sales.ByMonthlySubscription().ByMonth(previousMonth.Year(), previousMonth.Month())
	currentMonthSales := sales.ByMonth(m.Date.Year(), m.Date.Month())
	currentYearSales := sales.ByYear(m.Date.Year())

	// Basics
	m.Customers = len(currentMonthSales.CustomersMap())
	m.CustomersAnnual = len(currentMonthSales.ByAnnualSubscription().CustomersMap())
	m.CustomersMonthly = len(currentMonthSales.ByMonthlySubscription().CustomersMap())

	// New customers
	newCustomers := currentMonthSales.ByNewCustomers(allPreviousSales, m.Date)
	m.NewCustomersMonthly = len(newCustomers.ByMonthlySubscription().CustomersMap())
	m.NewCustomersAnnual = len(newCustomers.ByAnnualSubscription().CustomersMap())
	m.NewCustomers = len(newCustomers.CustomersMap())

	// Sales
	m.TotalSalesUSD.Total = currentMonthSales.TotalSumUSD()
	m.TotalSalesUSD.Fee = currentMonthSales.FeeSumUSD()

	// churn, no churn for first month
	m.HasChurnRate = m.PrevMonth != nil
	if m.HasChurnRate {
		nextMonth := m.Date.AddDate(0, 1, 0)
		nextMonthSales := sales.ByMonth(nextMonth.Year(), nextMonth.Month())

		m.ChurnRate, m.ChurnedCustomers, _ = computeMonthlyChurn(m.Date, previousMonthSales, currentMonthSales, nextMonthSales)
		m.ChurnRatePercentage = m.ChurnRate * 100
	}

	// Active customers
	if previousMonthData != nil {
		m.ActiveCustomersMonthly = previousMonthData.ActiveCustomersMonthly + m.NewCustomersMonthly - len(m.ChurnedCustomers)
		m.ActiveCustomersAnnual = previousMonthData.ActiveCustomersAnnual + m.NewCustomersAnnual
		m.ActiveCustomersTotal = previousMonthData.ActiveCustomersTotal + m.NewCustomers - len(m.ChurnedCustomers)
	} else {
		m.ActiveCustomersMonthly = m.NewCustomersMonthly
		m.ActiveCustomersAnnual = m.NewCustomersAnnual
		m.ActiveCustomersTotal = m.NewCustomers
	}

	// projected annual revenue
	if len(currentYearSales) > 0 {
		for _, sale := range currentYearSales.Before(m.Date) {
			if sale.Period == marketplace.AnnualSubscription {
				m.AnnualRevenueUSD += sale.AmountUSD
			}
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
				m.AnnualRevenueUSD += sale.AmountUSD * 12 * marketplace.Amount(monthSalesFactor)
			} else if sale.Period == marketplace.AnnualSubscription {
				m.AnnualRevenueUSD += sale.AmountUSD * marketplace.Amount(monthSalesFactor)
			}
		}

		monthsWithSales := m.Date.In(marketplace.ServerTimeZone).Month() - currentYearSales[0].Date.Month()
		projectionFactor := 1.0 + float64(12.0-monthsWithSales)/12.0
		// fixme don't use 80%, but 100%, for 4th year and later
		m.AnnualRevenueUSD = m.AnnualRevenueUSD * marketplace.Amount(projectionFactor) * 0.8
	}
}

func computeMonthlyChurn(date time.Time, previous marketplace.Sales, allCurrentMonth marketplace.Sales, allNextMonth marketplace.Sales) (float64, []marketplace.Customer, int) {
	if len(previous) == 0 {
		return 0.0, nil, 0
	}

	previousMonthCustomers := previous.CustomerSalesMap()
	graceTimeEnd := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, marketplace.ServerTimeZone).AddDate(0, 1, 3)

	graceTimeSales := append(allCurrentMonth, allNextMonth...).Before(graceTimeEnd)
	graceTimeMonthlySales := graceTimeSales.ByMonthlySubscription().CustomerSalesMap()
	graceTimeAnnualSales := graceTimeSales.ByAnnualSubscription().CustomerSalesMap()

	// handling the active, unfinished month
	now := time.Now().In(marketplace.ServerTimeZone)
	nowYear, nowMonth, _ := now.Date()
	currentMonthRef := now.AddDate(0, 0, -3)
	isActiveMonth := date.Year() == nowYear && date.Month() == nowMonth

	// collect customers, which bought in the previous, but not in current month
	var churned []marketplace.Customer
	for id, candidate := range previousMonthCustomers {
		_, ok := graceTimeMonthlySales[id]
		_, upgraded := graceTimeAnnualSales[id]
		paymentExpected := !isActiveMonth || candidate.LatestPurchase().AddDate(0, 1, 0).Before(currentMonthRef)

		if !ok && !upgraded && paymentExpected {
			// make sure that the user did not upgrade to yearly in this month
			churned = append(churned, candidate.Customer)
		}
	}

	totalUsers := len(previousMonthCustomers)
	return float64(len(churned)) / float64(totalUsers), churned, totalUsers
}
