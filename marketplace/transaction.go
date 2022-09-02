package marketplace

import (
	"sort"
	"strings"
	"time"
)

// Transaction contains basic information about a transaction
type Transaction struct {
	// ReferenceID is a unique ID of this transaction
	ReferenceID string `json:"ref"`
	// Date, depends on source. Either the day, when the sale was mode or when the trial expires.
	Date YearMonthDay `json:"date"`
	// Customer defines the customer, who paid for the license
	Customer Customer `json:"customer"`
}

type Transactions []Transaction

type CustomerTransactionMap map[CustomerID]Transactions

func NewTransaction(refId string, date YearMonthDay, customer Customer) Transaction {
	return Transaction{
		ReferenceID: refId,
		Date:        date,
		Customer:    customer,
	}
}

func (t Transactions) SortedByDate() Transactions {
	c := t
	sort.SliceStable(c, func(i, j int) bool {
		return !c[i].Date.IsAfter(c[j].Date)
	})
	return c
}

// FilterBy returns a new Sales slice, which contains all items, were the keep function returned true
func (t Transactions) FilterBy(keep func(transaction Transaction) bool) Transactions {
	var filtered Transactions
	for _, i := range t {
		if keep(i) {
			filtered = append(filtered, i)
		}
	}
	return filtered
}

func (t Transactions) Before(date time.Time) Transactions {
	return t.FilterBy(func(transaction Transaction) bool {
		return transaction.Date.AsDate().Before(date)
	})
}

// ByYearMonth returns a new Transactions slice, which contains all items bought in the particular month
func (t Transactions) ByYearMonth(month YearMonth) Transactions {
	return t.FilterBy(func(transactions Transaction) bool {
		date := transactions.Date
		return date.Year() == month.Year() && date.Month() == month.Month()
	})
}

func (t Transactions) GroupByCustomer() CustomerTransactionMap {
	mapping := make(CustomerTransactionMap)
	for _, transaction := range t {
		existing, _ := mapping[transaction.Customer.ID]
		mapping[transaction.Customer.ID] = append(existing, transaction)
	}
	return mapping
}

func (t CustomerTransactionMap) RetainCustomers(retained []Customer) CustomerTransactionMap {
	result := make(CustomerTransactionMap)
	for _, customer := range retained {
		if transactions, found := t[customer.ID]; found {
			result[customer.ID] = transactions
		}
	}
	return result
}

// DateGroupedTransactions defines a basic transaction tied to a date
type DateGroupedTransactions struct {
	Date         YearMonthDay
	Transactions Transactions
}

// GroupByDate groups transactions by date
func (t Transactions) GroupByDate(newestDateFirst bool) []DateGroupedTransactions {
	groups := make(map[YearMonthDay]Transactions)
	for _, sale := range t {
		values := groups[sale.Date]
		groups[sale.Date] = append(values, sale)
	}

	var groupedTransactions []DateGroupedTransactions
	for date, transactions := range groups {
		groupedTransactions = append(
			groupedTransactions,
			DateGroupedTransactions{
				Date:         date,
				Transactions: transactions,
			},
		)
	}

	sort.SliceStable(groupedTransactions, func(i, j int) bool {
		a := groupedTransactions[i]
		b := groupedTransactions[j]

		if newestDateFirst {
			return a.Date.IsAfter(b.Date)
		}
		return !a.Date.IsAfter(b.Date)
	})

	return groupedTransactions
}

// GroupByCountry returns transactions, grouped by customer's country. The result is sorted by number of transactions
func (t Transactions) GroupByCountry() []CountryTransactions {
	mapping := make(map[string]Transactions)
	for _, transaction := range t {
		existing, _ := mapping[transaction.Customer.Country]
		mapping[transaction.Customer.Country] = append(existing, transaction)
	}

	var result []CountryTransactions
	for country, transactions := range mapping {
		result = append(result, CountryTransactions{Country: country, Transactions: transactions})
	}
	sort.SliceStable(result, func(i, j int) bool {
		countA := len(result[i].Transactions)
		countB := len(result[j].Transactions)
		if countA == countB {
			return strings.Compare(result[i].Country, result[j].Country) < 0
		}
		return countA > countB
	})
	return result
}

type CountryTransactions struct {
	Country      string
	Transactions Transactions
}
