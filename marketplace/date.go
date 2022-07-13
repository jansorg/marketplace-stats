package marketplace

import "time"

// YearMonthDay represents year, month, day
type YearMonthDay [3]int

func NewYearMonthDay(year, month, day int) YearMonthDay {
	return [3]int{year, month, day}
}

func NewYearMonthDayByDate(date time.Time) YearMonthDay {
	year, month, day := date.In(ServerTimeZone).Date()
	return [3]int{year, int(month), day}
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

func (d YearMonthDay) AsYearMonth() YearMonth {
	return NewYearMonth(d.Year(), d.Month())
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

func (d YearMonthDay) IsBefore(o YearMonthDay) bool {
	return d.AsDate().Before(o.AsDate())
}

func (d YearMonthDay) Equals(o YearMonthDay) bool {
	return d.Year() == o.Year() && d.Month() == o.Month() && d.Day() == o.Day()
}

func (d YearMonthDay) NextDay() YearMonthDay {
	y, m, day := d.AsDate().AddDate(0, 0, 1).Date()
	return NewYearMonthDay(y, int(m), day)
}

func (d YearMonthDay) AddDays(days int) YearMonthDay {
	y, m, day := d.AsDate().AddDate(0, 0, days).Date()
	return NewYearMonthDay(y, int(m), day)
}

func (d YearMonthDay) Add(years, months, days int) YearMonthDay {
	y, m, day := d.AsDate().AddDate(years, months, days).Date()
	return NewYearMonthDay(y, int(m), day)
}

/* dateRangesTo returns a slice of begin,end tuples covering all dat */
func (d YearMonthDay) dateRangesTo(last YearMonthDay, addYears, addMonths, addDays int) [][2]YearMonthDay {
	var result [][2]YearMonthDay

	begin := d
	for begin.IsBefore(last) {
		end := begin.Add(addYears, addMonths, addDays)
		if end.IsAfter(last) {
			end = last
		}
		result = append(result, [2]YearMonthDay{begin, end})
		begin = end.AddDays(1)
	}

	return result
}
