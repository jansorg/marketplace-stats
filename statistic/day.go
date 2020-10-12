package statistic

import (
	"github.com/jansorg/marketplace-stats/marketplace"
	"time"
)

func NewDay(date time.Time) *Day {
	return &Day{
		Date: date,
	}
}

// Day contains the statistics of sales on a particular day
type Day struct {
	Date          time.Time
	TotalSalesUSD AmountAndFee
	Sales         marketplace.Sales
}

func (d *Day) Name() string {
	return d.Date.Format("2006-01-02")
}

func (d *Day) IsToday() bool {
	year, month, day := time.Now().In(marketplace.ServerTimeZone).Date()
	year2, month2, day2 := d.Date.In(marketplace.ServerTimeZone).Date()
	return year == year2 && month == month2 && day == day2
}

func (d *Day) Update(sales marketplace.Sales) {
	daySales := sales.ByDay(d.Date)
	d.Sales = daySales
	d.TotalSalesUSD = AmountAndFee{
		Total: daySales.TotalSumUSD(),
		Fee:   daySales.FeeSumUSD(),
	}
}
