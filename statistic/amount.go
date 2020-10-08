package statistic

import "jansorg/marketplace-stats/marketplace"

type AmountAndFee struct {
    Total marketplace.Amount
    Fee   marketplace.Amount
}

func (a AmountAndFee) PaidOut() marketplace.Amount {
    return a.Total - a.Fee
}
