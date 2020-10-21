package marketplace

import "time"

type DownloadAndDate struct {
	Downloads int

	Year      int
	Month     time.Month
	Day       int
}

func (d DownloadAndDate) Date() YearMonthDay {
	return NewYearMonthDay(d.Year, int(d.Month), d.Day)
}

type DownloadMonthly struct {
	Year      int
	Month     time.Month
	Downloads int
}

func (d DownloadMonthly) Date() YearMonth {
	return NewYearMonth(d.Year, d.Month)
}

type Filter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type NameValuePair struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type DownloadData struct {
	Dimension string `json:"dimension"`
	Serie     []NameValuePair
}

type DownloadResponse struct {
	Measure    string `json:"measure"`
	Filters    []Filter
	Dimension1 string `json:"dim1"`
	Data       DownloadData
}
