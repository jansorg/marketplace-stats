package statistic

import (
	"fmt"
	"time"

	"github.com/jansorg/marketplace-stats/marketplace"
)

type Week struct {
	BeginDate     time.Time
	TotalSalesUSD AmountAndFee
	Sales         marketplace.Sales
	Days          []*Day
}

func NewWeekToday(timeZone *time.Location) *Week {
	year, week := time.Now().In(timeZone).ISOWeek()
	return NewWeek(firstDayOfISOWeek(year, week, marketplace.ServerTimeZone))
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

//https://stackoverflow.com/questions/18624177/go-unix-timestamp-for-first-day-of-the-week-from-iso-year-week
func firstDayOfISOWeek(year int, week int, timezone *time.Location) time.Time {
	date := time.Date(year, 0, 0, 0, 0, 0, 0, timezone)
	isoYear, isoWeek := date.ISOWeek()
	for date.Weekday() != time.Monday { // iterate back to Monday
		date = date.AddDate(0, 0, -1)
		isoYear, isoWeek = date.ISOWeek()
	}
	for isoYear < year { // iterate forward to the first day of the first week
		date = date.AddDate(0, 0, 1)
		isoYear, isoWeek = date.ISOWeek()
	}
	for isoWeek < week { // iterate forward to the first day of the given week
		date = date.AddDate(0, 0, 1)
		isoYear, isoWeek = date.ISOWeek()
	}
	return date
}