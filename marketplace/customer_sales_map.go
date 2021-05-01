package marketplace

import (
	"fmt"
	"sort"
	"strconv"
	"time"
)

type CustomerSalesMap map[CustomerID]*CustomerSales

func (m CustomerSalesMap) TotalSales(id CustomerID) Amount {
	data, ok := m[id]
	if !ok {
		return 0
	}
	return data.TotalUSD
}

func (m CustomerSalesMap) PaidMonths(id CustomerID, before time.Time) int {
	data, ok := m[id]
	if !ok {
		return 0
	}

	count := 0
	var lastDate YearMonthDay
	for _, item := range data.Sales.SortedByDate() {
		current := item.Date
		if current.AsDate().After(before) {
			break
		}

		if current.AsYearMonth() != lastDate.AsYearMonth() {
			count++
		}
		lastDate = item.Date
	}
	return count
}

func (m CustomerSalesMap) GroupByPaidMonths(customers Customers, before time.Time) []NumberedGroup {
	groups := make(map[int]int)

	for _, customer := range customers {
		paid := m.PaidMonths(customer.ID, before)
		if paid > 0 {
			groups[paid]++
		}
	}

	var result []NumberedGroup
	for paid, count := range groups {
		result = append(result, NumberedGroup{
			Name:  fmt.Sprintf("%d", paid),
			Value: count,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		a, _ := strconv.Atoi(result[i].Name)
		b, _ := strconv.Atoi(result[j].Name)
		return a <= b
	})
	return result
}
