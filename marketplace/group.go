package marketplace

// GroupedSales is a number of sales, which have a group name
type GroupedSales struct {
	Name     string
	TotalUSD Amount
	Sales    Sales
}

// GroupedCustomers is a name with a list of customers
type GroupedCustomers struct {
	Name      string
	Customers Customers
}

// DateGroupedSales is a number of sales, which have a group name and a date
type DateGroupedSales struct {
	Date     YearMonthDay
	Name     string
	TotalUSD Amount
	Sales    Sales
}
