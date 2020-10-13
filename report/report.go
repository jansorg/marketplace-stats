package report

//go:generate esc -o static.go -pkg report static

import (
	"html/template"
	"strings"
	"time"

	"github.com/jansorg/marketplace-stats/marketplace"
	"github.com/jansorg/marketplace-stats/statistic"
)

type HTMLReport struct {
	Date       time.Time
	PluginInfo marketplace.PluginInfo

	Sales                    marketplace.Sales
	Customers                marketplace.Customers
	CustomerCount            int
	AnnualSubscriptionCount  int
	MonthlySubscriptionCount int
	FreeSubscriptionCount    int

	Week              *statistic.Week
	Years             []*statistic.Year
	CustomerSales     []*marketplace.CustomerSales
	CountrySales      []*marketplace.CountrySales
	SubscriptionSales []marketplace.GroupedSales
	CustomerTypeSales []marketplace.GroupedSales
	CurrencySales     []*marketplace.CurrencySales
}

func NewReport(pluginInfo marketplace.PluginInfo, allSales marketplace.Sales, years []*statistic.Year) HTMLReport {
	customers := allSales.Customers()
	week := statistic.NewWeekToday(marketplace.ServerTimeZone)
	week.Update(allSales)

	return HTMLReport{
		Date:                     time.Now(),
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
		Years:                    years,
	}
}

func (r HTMLReport) Generate() (string, error) {
	funcMap := template.FuncMap{
		"addInt": func(a, b int) int {
			return a + b
		},
		"subInt": func(a, b int) int {
			return a - b
		},
	}

	templateString := FSMustString(false, "/static/report.gohtml")
	report, err := template.New("basic").Funcs(funcMap).Parse(templateString)
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
