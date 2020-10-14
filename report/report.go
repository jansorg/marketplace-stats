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

	Timeline *Timeline
}

func NewReport(pluginInfo marketplace.PluginInfo, allSales marketplace.Sales, client marketplace.Client) (*HTMLReport, error) {
	monthlyDownloadsUnique, err := client.DownloadsMonthly(true, "", "", "", "", "")
	if err != nil {
		return nil, err
	}

	monthlyDownloadsTotal, err := client.DownloadsMonthly(false, "", "", "", "", "")
	if err != nil {
		return nil, err
	}

	// iterate years
	var years []*statistic.Year
	if len(allSales) > 0 {
		firstDate := allSales[0].Date.AsDate()
		lastDate := allSales[len(allSales)-1].Date.AsDate().AddDate(0, 1, 0)
		year := firstDate

		var previousYearStats *statistic.Year
		for !year.After(lastDate) {
			yearStats := statistic.NewYear(year.Year())
			yearStats.Update(previousYearStats, allSales, monthlyDownloadsTotal, monthlyDownloadsUnique)

			years = append(years, yearStats)
			year = year.AddDate(1, 0, 0)
			previousYearStats = yearStats
		}
	}

	customers := allSales.Customers()
	week := statistic.NewWeekToday(marketplace.ServerTimeZone)
	week.Update(allSales)

	return &HTMLReport{
		Date:                     time.Now(),
		PluginInfo:               pluginInfo,
		Timeline:                 NewMonthlyTimeline(allSales, monthlyDownloadsUnique),
		Sales:                    allSales,
		Week:                     week,
		Years:                    years,
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
	}, nil
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
