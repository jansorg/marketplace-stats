package marketplace

import "sort"

// Customers defines a slice of Customer
type Customers []Customer

// SortByID returns a new slice, which is ascendingly sorted by customer id.
func (c Customers) SortByID() Customers {
	sorted := append(Customers{}, c...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ID < sorted[j].ID
	})
	return sorted
}
