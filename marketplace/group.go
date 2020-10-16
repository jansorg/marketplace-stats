package marketplace

// GroupedSales is a number of sales, which have a group name
type GroupedSales struct {
	Name     string
	TotalUSD Amount
	Sales    Sales
}

// DateGroupedSales is a number of sales, which have a group name and a date
type DateGroupedSales struct {
	Date     YearMonthDay
	Name     string
	TotalUSD Amount
	Sales    Sales
}
