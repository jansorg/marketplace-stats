package main

import "github.com/jansorg/marketplace-stats/marketplace"

type dummyClient struct {
	customers marketplace.Customers
	sales     marketplace.Sales
	downloads []marketplace.DownloadAndDate
}

func NewDummyClient(customerCount, salesCount, salesMonths int) *dummyClient {
	customers := randomCustomers(customerCount)
	sales := randomSales(salesCount, salesMonths, customers).SortedByDate()
	downloads := randomDownloads(sales)
	return &dummyClient{
		customers: customers,
		sales:     sales,
		downloads: downloads,
	}
}

func (d *dummyClient) GetCurrentPluginInfo() (marketplace.PluginInfo, error) {
	return marketplace.PluginInfo{
		ID:          42,
		Name:        "DemoPlugin",
		Description: "This is a demo plugin",
		Link:        "",
	}, nil
}

func (d *dummyClient) GetPluginInfo(id string) (marketplace.PluginInfo, error) {
	panic("not supp")
}

func (d *dummyClient) DownloadsMonthly(uniqueDownloads bool, channel, build, product, country, productCommonCode string) ([]marketplace.DownloadMonthly, error) {
	monthMap := make(map[marketplace.YearMonth][]marketplace.DownloadMonthly)
	for _, d := range d.downloads {
		existing := monthMap[d.Date().AsYearMonth()]
		monthMap[d.Date().AsYearMonth()] = append(existing, marketplace.DownloadMonthly{
			Year:      d.Year,
			Month:     d.Month,
			Downloads: d.Downloads,
		})
	}

	var months []marketplace.DownloadMonthly
	for _, v := range monthMap {
		months = append(months, v...)
	}
	return months, nil
}

func (d *dummyClient) DownloadsWeekly(uniqueDownloads bool, channel, build, product, country, productCommonCode string) ([]marketplace.DownloadAndDate, error) {
	panic("implement me")
}

func (d *dummyClient) DownloadsDaily(uniqueDownloads bool, channel, build, product, country, productCommonCode string) ([]marketplace.DownloadAndDate, error) {
	panic("implement me")
}

func (d *dummyClient) Downloads(period string, uniqueDownloads bool, channel, build, product, country, productCommonCode string) (marketplace.DownloadResponse, error) {
	panic("not supported")
}

func (d *dummyClient) GetAllSalesInfo() (marketplace.Sales, error) {
	return d.sales, nil
}

func (d *dummyClient) GetSalesInfo(beginDate, endDate marketplace.YearMonthDay) (marketplace.Sales, error) {
	return d.sales.ByDateRange(beginDate, endDate), nil
}

func (d *dummyClient) GetJSON(path string, params map[string]string, target interface{}) error {
	panic("not supported")
}
