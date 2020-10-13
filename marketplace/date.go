package marketplace

import "time"

// YearMonthDay represents year, month, day
type YearMonthDay [3]int

func NewYearMonthDay(year, month, day int) YearMonthDay {
	return [3]int{year, month, day}
}

func (d YearMonthDay) String() string {
	return d.AsDate().Format("2006-01-02")
}

func (d YearMonthDay) Year() int {
	return d[0]
}

func (d YearMonthDay) Month() time.Month {
	return time.Month(d[1])
}

func (d YearMonthDay) Day() int {
	return d[2]
}

func (d YearMonthDay) AsDate() time.Time {
	return time.Date(d[0], time.Month(d[1]), d[2], 0, 0, 0, 0, ServerTimeZone)
}

func (d YearMonthDay) IsAfter(o YearMonthDay) bool {
	return d.AsDate().After(o.AsDate())
}

func (d YearMonthDay) Equals(o YearMonthDay) bool {
	return d.Year() == o.Year() && d.Month() == o.Month() && d.Day() == o.Day()
}

func (d YearMonthDay) NextDay() YearMonthDay {
	y, m, day := d.AsDate().AddDate(0, 0, 1).Date()
	return NewYearMonthDay(y, int(m), day)
}
