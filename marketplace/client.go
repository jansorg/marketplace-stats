package marketplace

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	pluginID string
	token    string
	client   http.Client
	hostname string
}

func NewClient(pluginID, token string) *Client {
	return &Client{
		token:    token,
		pluginID: pluginID,
		hostname: "plugins.jetbrains.com",
		client:   http.Client{},
	}
}

func (c *Client) GetCurrentPluginInfo() (PluginInfo, error) {
	return c.GetPluginInfo(c.pluginID)
}

func (c *Client) GetPluginInfo(id string) (PluginInfo, error) {
	var plugin PluginInfo
	err := c.GetJSON(fmt.Sprintf("/api/plugins/%s", id), nil, &plugin)
	return plugin, err
}

func (c *Client) GetAllSalesInfo() ([]Sale, error) {
	y, m, d := time.Now().Date()
	return c.GetSalesInfo(NewYearMonthDay(2019, 06, 25), NewYearMonthDay(y, int(m), d))
}

func (c *Client) GetSalesInfo(beginDate, endDate YearMonthDay) ([]Sale, error) {
	params := map[string]string{
		"beginDate": beginDate.String(),
		"endDate":   endDate.String(),
	}

	var sales []Sale
	err := c.GetJSON(fmt.Sprintf("/api/marketplace/tempApi/plugin/%s/sales-info", c.pluginID), params, &sales)
	return sales, err
}

func (c *Client) GetJSON(path string, params map[string]string, target interface{}) error {
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
