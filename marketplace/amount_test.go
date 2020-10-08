package marketplace

import "testing"
import "github.com/stretchr/testify/require"

func TestAmount_Format(t *testing.T) {
	require.EqualValues(t, "2,345.67", Amount(2345.67).Format())
}
