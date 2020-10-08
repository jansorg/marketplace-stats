package marketplace

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBasics(t *testing.T) {
	cust := NewCustomer(1, "User 1", "Country Name", AccountTypeOrganization)
	s := NewSale("1", 2020, 10, 1, MonthlySubscription, cust, 100, "EUR", 200)
	require.EqualValues(t, 15.0, s.FeeAmount())
	require.EqualValues(t, 30.0, s.FeeAmountUSD())
	require.EqualValues(t, 2.0, s.ExchangeRate())
	require.False(t, s.IsFreeSubscription())
}

func TestInitialFee(t *testing.T) {
	cust := NewCustomer(1, "User 1", "Country Name", AccountTypeOrganization)
	s := NewSale("1", 2020, 1, 1, MonthlySubscription, cust, 100, "EUR", 200)
	require.EqualValues(t, 5.0, s.FeeAmount())
	require.EqualValues(t, 10.0, s.FeeAmountUSD())
}
