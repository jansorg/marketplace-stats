package marketplace

import "time"

func PaidOutPercentage(date time.Time) float64 {
	if date.Before(feeChangeDate) {
		return 0.95
	}
	return 0.85
}

func FeePercentage(date time.Time) float64 {
	if date.Before(feeChangeDate) {
		return 0.05
	}
	return 0.15
}
