package statistic

import (
    "fmt"
    "jansorg/marketplace-stats/marketplace"
)

// NewYear returns a pointer to a Year
func NewYear(year int) *Year {
    return &Year{
        Year: year,
    }
}

// Year contains the statistics for a given year
type Year struct {
    Year int

    TotalCustomers        int
    TotalCustomersAnnual  int
    TotalCustomersMonthly int
    TotalSalesUSD         AmountAndFee
}

func (y *Year) Name() string {
    return fmt.Sprintf("%d", y.Year)
}

func (y *Year) Update(sales marketplace.Sales) {
    yearlySales := sales.ByYear(y.Year)

    y.TotalCustomers = len(yearlySales.CustomersMap())
    y.TotalCustomersAnnual = len(yearlySales.ByAnnualSubscription().CustomersMap())
    y.TotalCustomersMonthly = len(yearlySales.ByMonthlySubscription().CustomersMap())
    y.TotalSalesUSD.Total = yearlySales.TotalSumUSD()
    y.TotalSalesUSD.Fee = yearlySales.FeeSumUSD()
}
