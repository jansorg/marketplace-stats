package marketplace

import (
	"sort"
	"strings"
	"time"
)

// Subscription is either an Annual or a Monthly subscription
type Subscription string

// AccountType is either a Personal or an Organization account
type AccountType string

// Sales is a slice of sales. It offers a wide range of methods to aggregate the data.
type Sales []Sale

const (
	AnnualSubscription  Subscription = "Annual"
	MonthlySubscription Subscription = "Monthly"

	AccountTypePersonal     AccountType = "Personal"
	AccountTypeOrganization AccountType = "Organization"
)

// FilterBy returns a new Sales slice, which contains all items, were the keep function returned true
func (s Sales) FilterBy(keep func(Sale) bool) Sales {
	var filtered Sales
	for _, s := range s {
		if keep(s) {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// ByDay returns a new Sales slice, which contains all items bought at this particular day
func (s Sales) ByDay(date time.Time) Sales {
	y, m, d := date.Date()
	return s.FilterBy(func(sale Sale) bool {
		date := sale.Date
		return date.Year() == y && date.Month() == m && date.Day() == d
	})
}

// ByYearMonthDay returns a new Sales slice, which contains all items bought at this particular day
func (s Sales) ByYearMonthDay(day YearMonthDay) Sales {
	return s.FilterBy(func(sale Sale) bool {
		date := sale.Date
		return date.Year() == day.Year() && date.Month() == day.Month() && date.Day() == day.Day()
	})
}

// ByYearMonth returns a new Sales slice, which contains all items bought at this particular day
func (s Sales) ByYearMonth(month YearMonth) Sales {
	return s.FilterBy(func(sale Sale) bool {
		date := sale.Date
		return date.Year() == month.Year() && date.Month() == month.Month()
	})
}

// ByWeek returns a new Sales slice, which contains all items bought in the week of the year
func (s Sales) ByWeek(year int, isoWeek int) Sales {
	return s.FilterBy(func(sale Sale) bool {
		y, w := sale.Date.AsDate().ISOWeek()
		return year == y && isoWeek == w
	})
}

func (s Sales) ByYear(year int) Sales {
	return s.FilterBy(func(sale Sale) bool {
		return sale.Date.Year() == year
	})
}

func (s Sales) ByMonth(year int, month time.Month) Sales {
	return s.FilterBy(func(sale Sale) bool {
		return sale.Date.Year() == year && sale.Date.Month() == month
	})
}

func (s Sales) Before(date time.Time) Sales {
	return s.FilterBy(func(sale Sale) bool {
		return sale.Date.AsDate().Before(date)
	})
}

func (s Sales) After(date time.Time) Sales {
	return s.FilterBy(func(sale Sale) bool {
		return sale.Date.AsDate().After(date)
	})
}

func (s Sales) ByMonthlySubscription() Sales {
	return s.FilterBy(func(sale Sale) bool {
		return sale.Period == MonthlySubscription
	})
}

func (s Sales) ByAnnualSubscription() Sales {
	return s.FilterBy(func(sale Sale) bool {
		return sale.Period == AnnualSubscription
	})
}

func (s Sales) ByFreeSubscription() Sales {
	return s.FilterBy(Sale.IsFreeSubscription)
}

func (s Sales) ByCustomer(c Customer) Sales {
	return s.FilterBy(func(sale Sale) bool {
		return sale.Customer.ID == c.ID
	})
}

func (s Sales) ByAccountType(subscription AccountType) Sales {
	return s.FilterBy(func(sale Sale) bool {
		return sale.Customer.Type == subscription
	})
}

func (s Sales) ByNewCustomers(allPreviousSales Sales, referenceDate time.Time) Sales {
	previousCustomers := allPreviousSales.Before(referenceDate).CustomersMap()
	return s.FilterBy(func(sale Sale) bool {
		_, seen := previousCustomers[sale.Customer.ID]
		return !seen
	})
}

func (s Sales) CustomersMap() map[CustomerID]Customer {
	result := make(map[CustomerID]Customer)
	for _, s := range s {
		customer := s.Customer
		_, ok := result[customer.ID]
		if !ok {
			result[customer.ID] = customer
		}
	}
	return result
}

func (s Sales) Customers() Customers {
	var result Customers
	for _, s := range s {
		result = append(result, s.Customer)
	}
	return result
}

func (s Sales) TotalSumUSD() Amount {
	var sum Amount
	for _, s := range s {
		sum += s.AmountUSD
	}
	return sum
}

func (s Sales) FeeSumUSD() Amount {
	var sum Amount
	for _, s := range s {
		sum += s.FeeAmountUSD()
	}
	return sum
}

func (s Sales) PaidOutUSD() Amount {
	return s.TotalSumUSD() - s.FeeSumUSD()
}

func (s Sales) CustomerSalesMap() map[CustomerID]*CustomerSales {
	mapping := make(map[CustomerID]*CustomerSales)
	for _, sale := range s {
		value, seen := mapping[sale.Customer.ID]
		if !seen {
			value = &CustomerSales{
				Customer: sale.Customer,
				Sales:    Sales{},
				TotalUSD: 0.0,
			}
		}
		value.Sales = append(value.Sales, sale)
		value.TotalUSD += sale.AmountUSD
		mapping[sale.Customer.ID] = value
	}
	return mapping
}

func (s Sales) CustomerSales() []*CustomerSales {
	mapping := s.CustomerSalesMap()

	var result []*CustomerSales
	for _, v := range mapping {
		result = append(result, v)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].TotalUSD == result[j].TotalUSD {
			return strings.Compare(result[i].Customer.Name, result[j].Customer.Name) < 0
		}
		return result[i].TotalUSD > result[j].TotalUSD
	})
	return result
}

func (s Sales) CountrySales() []*CountrySales {
	mapping := make(map[string]*CountrySales)
	for _, sale := range s {
		value, seen := mapping[sale.Customer.Country]
		if !seen {
			value = &CountrySales{
				Country:  sale.Customer.Country,
				Sales:    Sales{},
				TotalUSD: 0.0,
			}
		}
		value.Sales = append(value.Sales, sale)
		value.TotalUSD += sale.AmountUSD
		mapping[sale.Customer.Country] = value
	}

	var result []*CountrySales
	for _, v := range mapping {
		result = append(result, v)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].TotalUSD == result[j].TotalUSD {
			return strings.Compare(result[i].Country, result[j].Country) < 0
		}
		return result[i].TotalUSD > result[j].TotalUSD
	})
	return result
}

func (s Sales) SubscriptionSales() []GroupedSales {
	annual := s.ByAnnualSubscription()
	monthly := s.ByMonthlySubscription()
	result := []GroupedSales{
		{
			Name:     "Annual",
			TotalUSD: annual.TotalSumUSD(),
			Sales:    annual,
		},
		{
			Name:     "Monthly",
			TotalUSD: monthly.TotalSumUSD(),
			Sales:    monthly,
		},
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalUSD > result[j].TotalUSD
	})
	return result
}

func (s Sales) CustomerTypeSales() []GroupedSales {
	organizations := s.ByAccountType(AccountTypeOrganization)
	persons := s.ByAccountType(AccountTypePersonal)
	result := []GroupedSales{
		{
			Name:     "Organization",
			TotalUSD: organizations.TotalSumUSD(),
			Sales:    organizations,
		},
		{
			Name:     "Person",
			TotalUSD: persons.TotalSumUSD(),
			Sales:    persons,
		},
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalUSD > result[j].TotalUSD
	})
	return result
}

// sales in currencies, sorted by USD
func (s Sales) GroupByCurrency() []*CurrencySales {
	mapping := make(map[Currency]*CurrencySales)

	for _, sale := range s {
		value, seen := mapping[sale.Currency]
		if !seen {
			value = &CurrencySales{
				Currency: sale.Currency,
			}
			mapping[sale.Currency] = value
		}
		value.TotalSales += sale.Amount
		value.TotalSalesUSD += sale.AmountUSD
	}

	var result []*CurrencySales
	for _, v := range mapping {
		result = append(result, v)
	}
	sort.Slice(result, func(i, j int) bool {
		a := result[i].TotalSalesUSD
		b := result[j].TotalSalesUSD
		if a == b {
			return strings.Compare(string(result[i].Currency), string(result[j].Currency)) < 0
		}
		return a > b
	})
	return result
}
