package statistic

import (
	"fmt"
	"time"

	"github.com/jansorg/marketplace-stats/marketplace"
)

// NewYear returns a pointer to a Year
func NewYear(year int) *Year {
	return &Year{
		Year: year,
	}
}

// Year contains the statistics for a given year
type Year struct {
	Year   int
	Months []*Month

	TotalCustomers        int
	TotalCustomersAnnual  int
	TotalCustomersMonthly int
	TotalSalesUSD         AmountAndFee

	DownloadsTotal  int
	DownloadsUnique int
}

func (y *Year) Name() string {
	return fmt.Sprintf("%d", y.Year)
}

func (y *Year) LastMonth() *Month {
	if len(y.Months) == 0 {
		return nil
	}
	return y.Months[len(y.Months)-1]
}

func (y *Year) Update(previousYear *Year, sales marketplace.Sales, trials marketplace.Transactions, downloadsTotal, downloadsUnique []marketplace.DownloadMonthly, graceDays int) {
	yearlySales := sales.ByYear(y.Year)

	y.TotalCustomers = len(yearlySales.CustomersMap())
	y.TotalCustomersAnnual = len(yearlySales.ByAnnualSubscription().CustomersMap())
	y.TotalCustomersMonthly = len(yearlySales.ByMonthlySubscription().CustomersMap())
	y.TotalSalesUSD.Total = yearlySales.TotalSumUSD()
	y.TotalSalesUSD.Fee = yearlySales.FeeSumUSD()

	// iterate months
	if len(sales) > 0 {
		currentMonth := time.Date(y.Year, time.January, 1, 0, 0, 0, 0, marketplace.ServerTimeZone)
		lastMonth := time.Date(y.Year, time.December, 30, 23, 59, 59, 999, marketplace.ServerTimeZone)

		if len(sales) > 0 && sales[0].Date.AsDate().After(currentMonth) {
			currentMonth = time.Date(sales[0].Date.Year(), sales[0].Date.Month(), 1, 0, 0, 0, 0, marketplace.ServerTimeZone)
		}

		now := time.Now().In(marketplace.ServerTimeZone)
		if lastMonth.After(now) {
			lastMonth = now.AddDate(0, 1, -now.Day())
		}

		var prevMonthData *Month
		if previousYear != nil {
			prevMonthData = previousYear.LastMonth()
		}

		for !currentMonth.After(lastMonth) {
			month := NewMonthForDate(currentMonth)
			month.Update(sales, trials, prevMonthData, downloadsTotal, downloadsUnique, graceDays)

			y.Months = append(y.Months, month)
			currentMonth = currentMonth.AddDate(0, 1, 0)
			prevMonthData = month
		}

		// calculate total downloads
		for _, d := range downloadsTotal {
			if d.Year == y.Year {
				y.DownloadsTotal += d.Downloads
			}
		}
		for _, d := range downloadsUnique {
			if d.Year == y.Year {
				y.DownloadsUnique += d.Downloads
			}
		}
	}
}
