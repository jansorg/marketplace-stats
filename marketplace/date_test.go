package marketplace

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDate(t *testing.T) {
	require.EqualValues(t, "2020-02-03", NewYearMonthDay(2020, 2, 3).String())

	require.True(t, NewYearMonthDay(2020, 2, 2).IsAfter(NewYearMonthDay(2020, 2, 1)))
	require.False(t, NewYearMonthDay(2020, 2, 2).IsAfter(NewYearMonthDay(2020, 2, 2)))
	require.False(t, NewYearMonthDay(2020, 2, 2).IsAfter(NewYearMonthDay(2020, 2, 3)))
}
