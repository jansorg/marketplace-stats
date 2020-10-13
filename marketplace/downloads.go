package marketplace

import "time"

type DownloadsDaily struct {
	Year      int
	Month     time.Month
	Day       int
	Downloads int
}

func (d DownloadsDaily) Date() YearMonthDay {
	return NewYearMonthDay(d.Year, int(d.Month), d.Day)
}

type DownloadsMonthly struct {
	Year      int
	Month     time.Month
	Downloads int
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
	Measure    string `json:"downloads-unique"`
	Filters    []Filter
	Dimension1 string `json:"dim1"`
	Data       DownloadData
}
