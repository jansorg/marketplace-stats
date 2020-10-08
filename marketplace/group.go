package marketplace

// GroupedSales is a number of sales, which have a group name
type GroupedSales struct {
	Name     string
	TotalUSD Amount
	Sales    Sales
}
