package marketplace

import (
	"sort"
)

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

// SortByID returns a slice sorted by the date of first purchase
func (c Customers) SortByDateMapping(mapping CustomerDateMap) Customers {
	sorted := append(Customers{}, c...)
	sort.Slice(sorted, func(i, j int) bool {
		return mapping[sorted[i].ID].IsBefore(mapping[sorted[j].ID])
	})
	return sorted
}

func (c Customers) GroupByCountry() []GroupedCustomers {
	countries := make(map[string]*GroupedCustomers)
	for _, customer := range c {
		if entry, ok := countries[customer.Country]; !ok {
			countries[customer.Country] = &GroupedCustomers{Name: customer.Country, Customers: Customers{customer}}
		} else {
			entry.Customers = append(entry.Customers, customer)
			countries[customer.Country] = entry
		}
	}

	var result []GroupedCustomers
	for _, group := range countries {
		result = append(result, *group)
	}
	sort.Slice(result, func(i, j int) bool {
		return len(result[i].Customers) > len(result[j].Customers)
	})
	return result
}
