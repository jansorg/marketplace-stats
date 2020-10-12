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
}

func (y *Year) Name() string {
	return fmt.Sprintf("%d", y.Year)
}

func (y *Year) Update(sales marketplace.Sales, downloadsTotal, downloadsUnique []marketplace.DownloadsMonthly) {
	yearlySales := sales.ByYear(y.Year)

	y.TotalCustomers = len(yearlySales.CustomersMap())
	y.TotalCustomersAnnual = len(yearlySales.ByAnnualSubscription().CustomersMap())
	y.TotalCustomersMonthly = len(yearlySales.ByMonthlySubscription().CustomersMap())
	y.TotalSalesUSD.Total = yearlySales.TotalSumUSD()
	y.TotalSalesUSD.Fee = yearlySales.FeeSumUSD()

	// iterate months
	if len(sales) > 0 {
		currentMonth := time.Date(y.Year, time.January, 1, 0, 0, 0, 0, marketplace.ServerTimeZone)
		if len(sales) > 0 && sales[0].Date.AsDate().After(currentMonth) {
			currentMonth = time.Date(sales[0].Date.Year(), sales[0].Date.Month(), 1, 0, 0, 0, 0, marketplace.ServerTimeZone)
		}

		now := time.Now().In(marketplace.ServerTimeZone)
		end := currentMonth.AddDate(1, 0, 0)
		if end.After(now) {
			end = now.AddDate(0, 1, -now.Day())
		}

		var prevMonthData *Month
		for !currentMonth.After(end) {
			month := NewMonthForDate(currentMonth)
			month.Update(sales, prevMonthData, downloadsTotal, downloadsUnique)

			y.Months = append(y.Months, month)
			currentMonth = currentMonth.AddDate(0, 1, 0)
			prevMonthData = month
		}
	}
}
