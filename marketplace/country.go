package marketplace

// CountrySales contains the sales of a specific country.
type CountrySales struct {
	Country  string
	TotalUSD Amount
	Sales    Sales
}
