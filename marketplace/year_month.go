package marketplace

import (
	"fmt"
	"time"
)

// YearMonth represents year, month
type YearMonth [2]int

func NewYearMonth(year int, month time.Month) YearMonth {
	return [2]int{year, int(month)}
}

func NewYearMonthByDate(date time.Time, timezone *time.Location) YearMonth {
	year, month, _ := date.In(timezone).Date()
	return [2]int{year, int(month)}
}

func (y YearMonth) Year() int {
	return y[0]
}

func (y YearMonth) String() string {
	return fmt.Sprintf("%d-%d", y.Year(), y.Month())
}

func (y YearMonth) Month() time.Month {
	return time.Month(y[1])
}

func (y YearMonth) AsDate() time.Time {
	return time.Date(y.Year(), y.Month(), 1, 0, 0, 0, 0, ServerTimeZone)
}

func (y YearMonth) IsAfter(o YearMonth) bool {
	return y.AsDate().After(o.AsDate())
}

func (y YearMonth) Equals(o YearMonth) bool {
	return y.Year() == o.Year() && y.Month() == o.Month()
}

func (y YearMonth) NextMonth() YearMonth {
	year, m, _ := y.AsDate().AddDate(0, 1, 0).Date()
	return NewYearMonth(year, m)
}

func (y YearMonth) PreviousMonth() YearMonth {
	year, m, _ := y.AsDate().AddDate(0, -1, 0).Date()
	return NewYearMonth(year, m)
}

func (y YearMonth) ContainsDate(date time.Time) bool {
	return y.Equals(NewYearMonthByDate(date, ServerTimeZone))
}
