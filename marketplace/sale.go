package marketplace

import (
	"fmt"
)

// Currency is a currency, as used by the JetBrains API
type Currency string

func NewSale(refID string, year, month, day int, subscription Subscription, customer Customer, amount Amount, currency Currency, amountUSD Amount) Sale {
	return Sale{
		ReferenceID: refID,
		Date:        NewYearMonthDay(year, month, day),
		Period:      subscription,
		Customer:    customer,
		Amount:      amount,
		Currency:    currency,
		AmountUSD:   amountUSD,
	}
}

// Sale represents a single transaction. Its structure is defined by the JetBrains API
type Sale struct {
	// ReferenceID is a unique ID of this transaction
	ReferenceID string `json:"ref"`
	// Date is the day, when this sale was made
	Date YearMonthDay `json:"date"`
	// Amount is the amount paid by the customer, in currency 'Currency'
	Amount Amount `json:"amount"`
	// Currency is the currency of the transaction
	Currency Currency `json:"currency"`
	// AmountUSD is Amount, converted into USD. This value is returned by the JetBrains API.
	AmountUSD Amount `json:"amountUSD"`
	// Period defines, if the transaction was for a monthly or annual license
	Period Subscription `json:"period"`
	// Customer defines the customer, who paid for the license
	Customer Customer `json:"customer"`
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
