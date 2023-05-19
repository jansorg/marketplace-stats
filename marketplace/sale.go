package marketplace

import (
	"fmt"
)

// Currency is a currency, as used by the JetBrains API
type Currency string

type Reseller struct {
	Code    int    `json:"code"`
	Name    string `json:"name"`
	Country string `json:"country"`
	Type    string `json:"type"`
}

type SaleDiscountDescription struct {
	Description string  `json:"description"`
	Percent     float64 `json:"percent"`
}

type SaleSubscriptionDates struct {
	Start YearMonthDay `json:"start"`
	End   YearMonthDay `json:"end"`
}

type SaleLineItem struct {
	Amount            Amount                    `json:"amount"`
	AmountUSD         Amount                    `json:"amountUSD"`
	Type              string                    `json:"type"`
	Discounts         []SaleDiscountDescription `json:"discountDescriptions"`
	LicenseIds        []string                  `json:"licenseIds"`
	SubscriptionDates SaleSubscriptionDates     `json:"subscriptionDates"`
}

const (
	LineItemTypeNew   = "NEW"
	LineItemTypeRenew = "RENEW"
)

func NewSale(refID string, year, month, day int, subscription Subscription, customer Customer, amount Amount, currency Currency, amountUSD Amount) Sale {
	return Sale{
		Transaction: NewTransaction(refID, NewYearMonthDay(year, month, day), customer),
		Period:      subscription,
		Amount:      amount,
		Currency:    currency,
		AmountUSD:   amountUSD,
	}
}

// Sale represents a single transaction. Its structure is defined by the JetBrains API
type Sale struct {
	Transaction
	// Amount is the amount paid by the customer, in currency 'Currency'
	Amount Amount `json:"amount"`
	// Currency is the currency of the transaction
	Currency Currency `json:"currency"`
	// AmountUSD is Amount, converted into USD. This value is returned by the JetBrains API.
	AmountUSD Amount `json:"amountUSD"`
	// Period defines, if the transaction was for a monthly or annual license
	Period Subscription `json:"period"`
	// Reseller who sold the license to the customer
	Reseller *Reseller `json:"reseller"`
	// Individual licenses of this sale
	LineItems []SaleLineItem `json:"lineItems"`
}

// ExchangeRate returns the exchange rate of AmountUSD / Amount
func (s Sale) ExchangeRate() float32 {
	return float32(s.AmountUSD) / float32(s.Amount)
}

// IsFreeSubscription returns true if the amount was 0, which indicates a free license
func (s Sale) IsFreeSubscription() bool {
	return s.Amount == 0.0
}

// FeeAmount returns the fee, in the currency of this sale, which is paid to JetBrains.
func (s Sale) FeeAmount() Amount {
	if s.Date.AsDate().Before(feeChangeDate) {
		return s.Amount * 0.05
	}
	return s.Amount * 0.15
}

// FeeAmountUSD is the fee in USD, which is paid to JetBrains.
func (s Sale) FeeAmountUSD() Amount {
	return s.AmountUSD * Amount(FeePercentage(s.Date.AsDate()))
}

func (s Sale) String() string {
	return fmt.Sprintf("[%v] %.2f USD, %s, customer %d", s.Date, s.AmountUSD, s.Period, s.Customer.ID)
}
