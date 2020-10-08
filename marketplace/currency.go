package marketplace

// CurrencySales contains the sales in a certain currency.
type CurrencySales struct {
	// Currency is the currency used in this struct
	Currency Currency
	// TotalSales is the sum of all sales in currency "Currency"
	TotalSales Amount
	// TotalSalesUSD is the sum of all sales, but in USD
	TotalSalesUSD Amount
	// Sales
	Sales Sales
}
