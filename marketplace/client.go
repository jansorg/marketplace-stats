package marketplace

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/sync/errgroup"
)

type Client interface {
	GetCurrentPluginInfo() (PluginInfo, error)
	GetPluginInfo(id string) (PluginInfo, error)

	GetCurrentPluginRating() (Rating, error)
	GetPluginRating(id string) (Rating, error)

	DownloadsMonthly(uniqueDownloads bool, channel, build, product, country, productCommonCode string) ([]DownloadMonthly, error)
	DownloadsWeekly(uniqueDownloads bool, channel, build, product, country, productCommonCode string) ([]DownloadAndDate, error)
	DownloadsDaily(uniqueDownloads bool, channel, build, product, country, productCommonCode string) ([]DownloadAndDate, error)
	Downloads(period string, uniqueDownloads bool, channel, build, product, country, productCommonCode string) (DownloadResponse, error)

	// GetAllSalesInfo returns the sales of the complete lifespan of the plugin
	GetAllSalesInfo() (Sales, error)
	// GetSalesInfo returns the sales in the given range.
	// If it covers more than one year, then the sales will be fetched in yearly chunks.
	GetSalesInfo(beginDate, endDate YearMonthDay) (Sales, error)

	// GetAllTrialsInfo returns the trial license data of the complete lifespan of the plugin
	GetAllTrialsInfo() (Transactions, error)
	// GetTrialsInfo returns the trial license data of the given range.
	// If it covers more than one year, then the sales will be fetched in yearly chunks.
	GetTrialsInfo(beginDate, endDate YearMonthDay) (Transactions, error)

	// GetJSON performs a generic request to the marketplace, it expects that JSON data is returned
	GetJSON(path string, params map[string]string, target interface{}) error
}

type client struct {
	pluginID string
	token    string
	client   http.Client
	hostname string
}

func NewClient(pluginID, token string) Client {
	return &client{
		token:    token,
		pluginID: pluginID,
		hostname: "plugins.jetbrains.com",
		client:   http.Client{},
	}
}

func (c *client) GetCurrentPluginInfo() (PluginInfo, error) {
	return c.GetPluginInfo(c.pluginID)
}

func (c *client) GetPluginInfo(id string) (PluginInfo, error) {
	var plugin PluginInfo
	err := c.GetJSON(fmt.Sprintf("/api/plugins/%s", id), nil, &plugin)
	return plugin, err
}

func (c *client) GetCurrentPluginRating() (Rating, error) {
	return c.GetPluginRating(c.pluginID)
}

func (c *client) GetPluginRating(id string) (Rating, error) {
	var rating Rating
	err := c.GetJSON(fmt.Sprintf("/api/plugins/%s/rating", id), nil, &rating)
	return rating, err
}

func (c *client) DownloadsMonthly(uniqueDownloads bool, channel, build, product, country, productCommonCode string) ([]DownloadMonthly, error) {
	resp, err := c.Downloads("month", uniqueDownloads, channel, build, product, country, productCommonCode)
	if err != nil {
		return nil, err
	}

	var months []DownloadMonthly
	for _, d := range resp.Data.Serie {
		parsedDate, err := time.ParseInLocation("2006-01-02", d.Name, ServerTimeZone)
		if err != nil {
			return nil, err
		}

		months = append(months, DownloadMonthly{
			Year:      parsedDate.Year(),
			Month:     parsedDate.Month(),
			Downloads: d.Value,
		})
	}
	return months, nil
}

func (c *client) DownloadsWeekly(uniqueDownloads bool, channel, build, product, country, productCommonCode string) ([]DownloadAndDate, error) {
	resp, err := c.Downloads("week", uniqueDownloads, channel, build, product, country, productCommonCode)
	if err != nil {
		return nil, err
	}

	var days []DownloadAndDate
	for _, d := range resp.Data.Serie {
		parsedDate, err := time.ParseInLocation("2006-01-02", d.Name, ServerTimeZone)
		if err != nil {
			return nil, err
		}

		days = append(days, DownloadAndDate{
			Year:      parsedDate.Year(),
			Month:     parsedDate.Month(),
			Day:       parsedDate.Day(),
			Downloads: d.Value,
		})
	}
	return days, nil
}

func (c *client) DownloadsDaily(uniqueDownloads bool, channel, build, product, country, productCommonCode string) ([]DownloadAndDate, error) {
	resp, err := c.Downloads("day", uniqueDownloads, channel, build, product, country, productCommonCode)
	if err != nil {
		return nil, err
	}

	var days []DownloadAndDate
	for _, d := range resp.Data.Serie {
		parsedDate, err := time.ParseInLocation("2006-01-02", d.Name, ServerTimeZone)
		if err != nil {
			return nil, err
		}

		days = append(days, DownloadAndDate{
			Year:      parsedDate.Year(),
			Month:     parsedDate.Month(),
			Day:       parsedDate.Day(),
			Downloads: d.Value,
		})
	}
	return days, nil
}

func (c *client) Downloads(period string, uniqueDownloads bool, channel, build, product, country, productCommonCode string) (DownloadResponse, error) {
	params := map[string]string{
		"plugin": c.pluginID,
	}

	if channel != "" {
		params["channel"] = channel
	}
	if build != "" {
		params["build"] = build
	}
	if product != "" {
		params["product"] = product
	}
	if country != "" {
		params["country"] = country
	}
	if productCommonCode != "" {
		params["product-common-code"] = productCommonCode
	}

	downloadType := "downloads-count"
	if uniqueDownloads {
		downloadType = "downloads-unique"
	}

	var resp DownloadResponse
	err := c.GetJSON(fmt.Sprintf("/statistic/%s/%s", downloadType, period), params, &resp)
	return resp, err
}

func (c *client) GetAllSalesInfo() (Sales, error) {
	y, m, d := time.Now().Date()
	return c.GetSalesInfo(NewYearMonthDay(2019, 06, 25), NewYearMonthDay(y, int(m), d))
}

func (c *client) GetSalesInfo(beginDate, endDate YearMonthDay) (Sales, error) {
	wg, _ := errgroup.WithContext(context.Background())

	// concurrently fetch sales of each year
	ranges := beginDate.dateRangesTo(endDate, 1, 0, 0)
	salesData := make([]Sales, len(ranges))
	for i, r := range ranges {
		i := i
		r := r
		wg.Go(func() error {
			var err error
			salesData[i], err = c.getSalesInfo(r[0], r[1])
			return err
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	var result Sales
	for _, data := range salesData {
		result = append(result, data...).SortedByDate()
	}
	return result, nil
}

func (c *client) GetAllTrialsInfo() (Transactions, error) {
	y, m, d := time.Now().Date()
	return c.GetTrialsInfo(NewYearMonthDay(2019, 06, 25), NewYearMonthDay(y, int(m), d))
}

func (c *client) GetTrialsInfo(beginDate, endDate YearMonthDay) (Transactions, error) {
	wg, _ := errgroup.WithContext(context.Background())

	// concurrently fetch sales of each year
	ranges := beginDate.dateRangesTo(endDate, 1, 0, 0)
	trialData := make([]Transactions, len(ranges))
	for i, r := range ranges {
		i := i
		r := r
		wg.Go(func() error {
			var err error
			trialData[i], err = c.getTrialInfo(r[0], r[1])
			return err
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	var result Transactions
	for _, data := range trialData {
		result = append(result, data...).SortedByDate()
	}
	return result, nil
}

func (c *client) GetJSON(path string, params map[string]string, target interface{}) error {
	u := url.URL{
		Scheme: "https",
		Host:   c.hostname,
		Path:   path,
	}

	q := u.Query()
	u.RawQuery = q.Encode()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status code %d: %s", resp.StatusCode, resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

// getSalesInfo fetches the sales data for the given range
// it must not be called with a range larger than one year - the remote API only allows one year at most
func (c *client) getSalesInfo(beginDate, endDate YearMonthDay) (Sales, error) {
	params := map[string]string{
		"beginDate": beginDate.String(),
		"endDate":   endDate.String(),
	}

	var sales []Sale
	err := c.GetJSON(fmt.Sprintf("/api/marketplace/plugin/%s/sales-info", c.pluginID), params, &sales)
	return sales, err
}

// getTrialInfo fetches the sales data for the given range
func (c *client) getTrialInfo(beginDate, endDate YearMonthDay) (Transactions, error) {
	params := map[string]string{
		"beginDate": beginDate.String(),
		"endDate":   endDate.String(),
	}

	var trialInfo []Transaction
	err := c.GetJSON(fmt.Sprintf("/api/marketplace/plugin/%s/trials-info", c.pluginID), params, &trialInfo)
	return trialInfo, err
}
