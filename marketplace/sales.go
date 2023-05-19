package marketplace

import (
	"sort"
	"strings"
	"time"
)

// Subscription is either an Annual or a Monthly subscription
type Subscription string

func (s Subscription) Abbrev() string {
	if s == AnnualSubscription {
		return "y"
	}
	if s == MonthlySubscription {
		return "m"
	}
	return "?"
}

// AccountType is either a Personal or an Organization account
type AccountType string

// Sales is a slice of sales. It offers a wide range of methods to aggregate the data.
type Sales []Sale

// CustomersMap maps a customer ID to customer
type CustomersMap map[CustomerID]Customer

// CustomerDateMap maps a customer ID to a YearMonthDate
type CustomerDateMap map[CustomerID]YearMonthDay

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

// ByYearMonth returns a new Sales slice, which contains all items bought in the particular month
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

// ByDateRange returns a new Sales slice, which contains all items bought in the given date range (inclusive)
func (s Sales) ByDateRange(begin, end YearMonthDay) Sales {
	return s.After(begin.AsDate()).Before(end.AsDate().AddDate(0, 0, 1))
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

func (s Sales) AtOrAfter(date time.Time) Sales {
	reference := NewYearMonthDayByDate(date)
	return s.FilterBy(func(sale Sale) bool {
		return sale.Date.Equals(reference) || sale.Date.IsAfter(reference)
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

func (s Sales) ByReturnedCustomers(previouslyChurned ChurnedCustomerList) Sales {
	churnedCustomers := previouslyChurned.Customers().AsMap()
	return s.FilterBy(func(sale Sale) bool {
		_, seen := churnedCustomers[sale.Customer.ID]
		return seen
	})
}

func (s Sales) CustomersMap() CustomersMap {
	result := make(CustomersMap)
	for _, s := range s {
		customer := s.Customer
		_, ok := result[customer.ID]
		if !ok {
			result[customer.ID] = customer
		}
	}
	return result
}

// Customers returns a unique set of customers found in the sales
func (s Sales) Customers() Customers {
	var result Customers
	for _, c := range s.CustomersMap() {
		result = append(result, c)
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

func (s Sales) CustomerCount() int {
	return len(s.CustomerSalesMap())
}

func (s Sales) CustomerSalesMap() CustomerSalesMap {
	mapping := make(CustomerSalesMap)
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

	sort.SliceStable(result, func(i, j int) bool {
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
	sort.SliceStable(result, func(i, j int) bool {
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

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].TotalUSD > result[j].TotalUSD
	})
	return result
}

func (s Sales) GroupByWeekday() []GroupedSales {
	result := make([]GroupedSales, 7)
	for day := time.Sunday; day <= time.Saturday; day++ {
		result[int(day)] = GroupedSales{
			Name: day.String(),
		}
	}

	for _, sale := range s {
		key := int(sale.Date.AsDate().Weekday())
		sales := append(result[key].Sales, sale)
		result[key].Sales = sales
		result[key].TotalUSD += sale.AmountUSD
	}
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

	sort.SliceStable(result, func(i, j int) bool {
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
	sort.SliceStable(result, func(i, j int) bool {
		a := result[i].TotalSalesUSD
		b := result[j].TotalSalesUSD
		if a == b {
			return strings.Compare(string(result[i].Currency), string(result[j].Currency)) < 0
		}
		return a > b
	})
	return result
}

// sales, grouped by year-month-day
func (s Sales) GroupByDate(newestDateFirst bool) []DateGroupedSales {
	groups := make(map[YearMonthDay]Sales)
	for _, sale := range s {
		values := groups[sale.Date]
		groups[sale.Date] = append(values, sale)
	}

	var groupedSales []DateGroupedSales
	for date, sales := range groups {
		groupedSales = append(
			groupedSales,
			DateGroupedSales{
				Date:     date,
				Name:     date.String(),
				TotalUSD: sales.TotalSumUSD(),
				Sales:    sales,
			},
		)
	}

	sort.SliceStable(groupedSales, func(i, j int) bool {
		a := groupedSales[i]
		b := groupedSales[j]

		if newestDateFirst {
			return a.Date.IsAfter(b.Date)
		}
		return !a.Date.IsAfter(b.Date)
	})

	return groupedSales
}

func (s Sales) SortedByDate() Sales {
	c := s
	sort.SliceStable(c, func(i, j int) bool {
		return !c[i].Date.IsAfter(c[j].Date)
	})
	return c
}

func (s Sales) Reversed() Sales {
	size := len(s)
	rev := make(Sales, size)
	for i, sale := range s {
		rev[size-i-1] = sale
	}
	return rev
}

func (s Sales) CustomersFirstPurchase() CustomerDateMap {
	result := make(CustomerDateMap)
	for _, sale := range s {
		stored, found := result[sale.Customer.ID]
		if !found || sale.Date.IsBefore(stored) {
			result[sale.Customer.ID] = sale.Date
		}
	}
	return result
}

func (s Sales) CustomersLastPurchase() CustomerDateMap {
	result := make(CustomerDateMap)
	for _, sale := range s {
		stored, found := result[sale.Customer.ID]
		if !found || sale.Date.IsAfter(stored) {
			result[sale.Customer.ID] = sale.Date
		}
	}
	return result
}

func (s Sales) UniqueLicenseCount() int {
	licenseIds := make(map[string]bool)

	for _, sale := range s {
		for _, items := range sale.LineItems {
			for _, licenseId := range items.LicenseIds {
				licenseIds[licenseId] = true
			}
		}
	}

	return len(licenseIds)
}

func (m CustomersMap) Without(customersMap CustomersMap) CustomersMap {
	result := make(CustomersMap)
	for k, v := range m {
		_, found := customersMap[k]
		if !found {
			result[k] = v
		}
	}
	return result
}
