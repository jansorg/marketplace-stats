package main

import (
	"fmt"
	"github.com/jansorg/marketplace-stats/marketplace"
	"github.com/jansorg/marketplace-stats/report"
)

func main() {
	c1 := marketplace.NewCustomer(123, "Customer 1", "Germany", marketplace.AccountTypePersonal)
	timeline := report.Timeline{
		Days: []report.TimelineDay{
			{
				Day:       marketplace.NewYearMonthDay(2020, 10, 1),
				Downloads: 10,
				Sales: marketplace.Sales{
					marketplace.NewSale("1", 2020, 10, 1, marketplace.AnnualSubscription, c1, marketplace.Amount(100.0), "USD", marketplace.Amount(10.0)),
				},
			},
			{
				Day:       marketplace.NewYearMonthDay(2020, 10, 2),
				Downloads: 30,
				Sales: marketplace.Sales{
					marketplace.NewSale("2", 2020, 10, 2, marketplace.AnnualSubscription, c1, marketplace.Amount(100.0), "USD", marketplace.Amount(128.0)),
				},
			},
			{
				Day:       marketplace.NewYearMonthDay(2020, 10, 3),
				Downloads: 90,
				Sales: marketplace.Sales{
					marketplace.NewSale("1", 2020, 10, 3, marketplace.AnnualSubscription, c1, marketplace.Amount(100.0), "USD", marketplace.Amount(68.0)),
				},
			},
		},
	}

	fmt.Println(timeline.DrawSVG())
}
