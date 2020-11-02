package marketplace

import (
	"fmt"
	"time"
)

// CustomerID identifies a specific customer, this is used by the JetBrains API
type CustomerID int

// CustomerSales contains the sales of a specific customer
type CustomerSales struct {
	Customer Customer
	Sales    Sales
	TotalUSD Amount
}

// LatestPurchase returns the latest purchase of this customer.
// If there are no sales, then the zero value of time.Time is returned.
func (c *CustomerSales) LatestPurchase() time.Time {
	var lastDate time.Time
	for _, s := range c.Sales {
		date := s.Date.AsDate()
		if date.After(lastDate) {
			lastDate = date
		}
	}
	return lastDate
}

// FirstPurchase returns the first purchase of this customer.
// If there are no sales, then the zero value of time.Time is returned.
func (c *CustomerSales) FirstPurchase() YearMonthDay {
	var firstDate YearMonthDay
	for _, s := range c.Sales {
		date := s.Date
		if date.IsBefore(firstDate) {
			firstDate = date
		}
	}
	return firstDate
}

// NewCustomer returns a new customer
func NewCustomer(id CustomerID, name, country string, accountType AccountType) Customer {
	return Customer{
		ID:      id,
		Name:    name,
		Country: country,
		Type:    accountType,
	}
}

// Customer defines a specific customer
type Customer struct {
	ID      CustomerID  `json:"code"`
	Name    string      `json:"name"`
	Country string      `json:"country"`
	Type    AccountType `json:"type"`
}

// String returns a string representation for debugging purposes
func (c Customer) String() string {
	return fmt.Sprintf("[%s] %s (%v), %s", c.Type, c.Name, c.ID, c.Country)
}
