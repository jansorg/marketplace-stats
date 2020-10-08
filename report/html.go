package report

//go:generate esc -o static.go -pkg report static

import (
	"html/template"
	"jansorg/marketplace-stats/marketplace"
	"jansorg/marketplace-stats/statistic"
	"strings"
)

type HTMLReport struct {
	PluginInfo marketplace.PluginInfo

	Sales                    marketplace.Sales
	Customers                marketplace.Customers
	CustomerCount            int
	AnnualSubscriptionCount  int
	MonthlySubscriptionCount int
	FreeSubscriptionCount    int

	Week              *statistic.Week
	Months            []*statistic.Month
	Years             []*statistic.Year
	CustomerSales     []*marketplace.CustomerSales
	CountrySales      []*marketplace.CountrySales
	SubscriptionSales []marketplace.GroupedSales
	CustomerTypeSales []marketplace.GroupedSales
	CurrencySales     []*marketplace.CurrencySales
}

func NewReport(pluginInfo marketplace.PluginInfo, allSales marketplace.Sales, years []*statistic.Year, months []*statistic.Month) HTMLReport {
	customers := allSales.Customers()
	week := statistic.NewWeekToday(marketplace.ServerTimeZone)
	week.Update(allSales)

	return HTMLReport{
		PluginInfo:               pluginInfo,
		Sales:                    allSales,
		Customers:                customers.SortByID(),
		CustomerSales:            allSales.CustomerSales(),
		CountrySales:             allSales.CountrySales(),
		SubscriptionSales:        allSales.SubscriptionSales(),
		CustomerTypeSales:        allSales.CustomerTypeSales(),
		CurrencySales:            allSales.GroupByCurrency(),
		CustomerCount:            len(customers),
		AnnualSubscriptionCount:  len(allSales.ByAnnualSubscription().Customers()),
		MonthlySubscriptionCount: len(allSales.ByMonthlySubscription().Customers()),
		FreeSubscriptionCount:    len(allSales.ByFreeSubscription().Customers()),
		Week:                     week,
		Months:                   months,
		Years:                    years,
	}
}

func (r HTMLReport) Generate() (string, error) {
	// language=HTML
	templateString := FSMustString(false, "/static/report.gohtml")
	report, err := template.New("basic").Parse(templateString)
	if err != nil {
		return "", err
	}

	w := strings.Builder{}
	err = report.Execute(&w, r)
	if err != nil {
		return "", err
	}
	return w.String(), nil
}
