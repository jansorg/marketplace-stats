package marketplace

import (
	"fmt"
	"sort"
	"strconv"
)

type ChurnedCustomer struct {
	Customer
	LastPurchase YearMonthDay
	Subscription Subscription
}

func (c ChurnedCustomer) PaidDuration(first YearMonthDay) string {
	var count int
	if c.Subscription == AnnualSubscription {
		count = c.LastPurchase.Year() - first.Year() + 1
	} else {
		count = 1
		for first.IsBefore(c.LastPurchase) {
			first = first.Add(0, 1, 0)
			count++
		}
	}
	return fmt.Sprintf("%d%s", count, c.Subscription.Abbrev())
}

type ChurnedCustomers []ChurnedCustomer

func (c ChurnedCustomers) Contains(id CustomerID) bool {
	for _, customer := range c {
		if id == customer.ID {
			return true
		}
	}
	return false
}

func (c ChurnedCustomers) SortedByDate() ChurnedCustomers {
	sorted := append(ChurnedCustomers{}, c...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].LastPurchase.IsBefore(sorted[j].LastPurchase)
	})
	return sorted
}

func (c ChurnedCustomers) Customers() Customers {
	var customers Customers
	for _, e := range c {
		customers = append(customers, e.Customer)
	}
	return customers
}

func (c ChurnedCustomers) GroupByPaidDuration(first CustomerDateMap) []NumberedGroup {
	mapping := make(map[string]int)
	for _, e := range c {
		paid := e.PaidDuration(first[e.ID])
		mapping[paid] = mapping[paid] + 1
	}

	var groups []NumberedGroup
	for k, v := range mapping {
		groups = append(groups, NumberedGroup{
			Name:  k,
			Value: v,
		})
	}
	sort.Slice(groups, func(i, j int) bool {
		a := groups[i]
		b := groups[j]
		if a.Name == b.Name {
			return a.Value > b.Value
		}

		paidA, _ := strconv.Atoi(a.Name[0 : len(a.Name)-1])
		paidB, _ := strconv.Atoi(b.Name[0 : len(b.Name)-1])
		if a.Name[len(a.Name)-1] == 'a' {
			paidA *= 12
		}
		if b.Name[len(b.Name)-1] == 'a' {
			paidB *= 12
		}
		return paidA > paidB
	})
	return groups
}
