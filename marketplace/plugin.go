package marketplace

import "strconv"

type PurchaseInfo struct {
	ProductCode string `json:"productCode"`
	// buyUrl
	// purchaseTerms
}

type Vendor struct {
	ID                  int      `json:"vendorId"`
	Name                string   `json:"name"`
	PublicName          string   `json:"publicName"`
	Email               string   `json:"email"`
	CountryCode         string   `json:"countryCode"`
	Country             string   `json:"country"`
	City                string   `json:"city"`
	ZipCode             string   `json:"zipCode"`
	URL                 string   `json:"url"`
	MarketplacePath     string   `json:"link"`
	TotalPlugins        int      `json:"totalPlugins"`
	TotalUsers          int      `json:"totalUsers"`
	Verified            bool     `json:"isVerified"`
	ServicesDescription []string `json:"servicesDescription"`
}

type Tag struct {
	ID              int    `json:"ID"`
	Name            string `json:"name"`
	Privileged      bool   `json:"privileged"`
	Searchable      bool   `json:"searchable"`
	MarketplacePath string `json:"link"`
}

type ScreenshotInfo struct {
	ID              int    `json:"id"`
	MarketplacePath string `json:"url"`
}

type Rating struct {
	MeanRating float64        `json:"meanRating"`
	MeanVotes  int            `json:"meanVotes"`
	UserRating int            `json:"userRating"`
	Votes      map[string]int `json:"votes"`
}

func (r Rating) CalculateRating() float64 {
	return (float64(r.VotesSum()) + 2.0*r.MeanRating) / (float64(r.VotesCount()) + 2.0)
}

func (r Rating) VotesCount() int {
	votesCount := 0
	for _, v := range r.Votes {
		votesCount += v
	}
	return votesCount
}

func (r Rating) VotesSum() int {
	votesSum := 0
	for k, v := range r.Votes {
		multiplier, _ := strconv.Atoi(k)
		votesSum += multiplier * v
	}
	return votesSum
}

type PluginInfo struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Link          string `json:"link"`
	Approved      bool   `json:"approve"`
	XmlID         string `json:"xmlId"`
	CustomIdeList bool   `json:"customIdeList"`
	PreviewText   string `json:"preview"`

	ContactEmail string `json:"email"`
	Copyright    string `json:"copyright"`
	Downloads    int    `json:"downloads"`
	PurchaseInfo PurchaseInfo
	//"cdate": 1601924313000,
	//"family": "intellij",

	URLs                map[string]string `json:"urls"`
	Tags                []Tag             `json:"tags"`
	RemovalRequested    bool              `json:"removalRequested"`
	HasUnapprovedUpdate bool              `json:"hasUnapprovedUpdate"`
	ReadyForSale        bool              `json:"readyForSale"`
	IconMarketplacePath string            `json:"icon"`
}
