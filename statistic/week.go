package statistic

import (
	"fmt"
	"time"

	"jansorg/marketplace-stats/marketplace"
)

type Week struct {
	BeginDate     time.Time
	TotalSalesUSD AmountAndFee
	Sales         marketplace.Sales
	Days          []*Day
}

func NewWeekToday(timeZone *time.Location) *Week {
	today := time.Now().In(timeZone)
	firstDay := today.AddDate(0, 0, -int(today.Weekday())+1)
	return NewWeek(firstDay)
}

func NewWeek(firstDay time.Time) *Week {
	begin := firstDay.Truncate(time.Hour)
	return &Week{
		BeginDate: begin,
	}
}

func (m *Week) Name() string {
	year, week := m.BeginDate.ISOWeek()
	return fmt.Sprintf("Week %d, %d", week, year)
}

// Update the current weeks's data from the complete collection of sales
func (m *Week) Update(sales marketplace.Sales) {
	weekSales := sales.ByWeek(m.BeginDate.ISOWeek())

	var days []*Day
	for i := 0; i < 7; i++ {
		day := NewDay(m.BeginDate.AddDate(0, 0, i))
		day.Update(sales)
		days = append(days, day)
	}

	m.Days = days
	m.Sales = weekSales
	m.TotalSalesUSD = AmountAndFee{
		Total: weekSales.TotalSumUSD(),
		Fee:   weekSales.FeeSumUSD(),
	}
}
