package report

//go:generate esc -o static.go -pkg report static

import (
	"fmt"
	"github.com/jansorg/marketplace-stats/util"
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
	WeekdaySales      []marketplace.GroupedSales

	Timeline *Timeline
}

func NewReport(pluginInfo marketplace.PluginInfo, allSalesUnsorted marketplace.Sales, client marketplace.Client) (*HTMLReport, error) {
	allSales := allSalesUnsorted.SortedByDate()

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

	week := statistic.NewWeekToday(marketplace.ServerTimeZone)
	week.Update(allSales)

	customers := allSales.Customers().SortByID()

	return &HTMLReport{
		Date:                     time.Now(),
		PluginInfo:               pluginInfo,
		Timeline:                 NewMonthlyTimeline(allSales, monthlyDownloadsUnique),
		Sales:                    allSales,
		Week:                     week,
		Years:                    years,
		Customers:                customers,
		CustomerSales:            allSales.CustomerSales(),
		CountrySales:             allSales.CountrySales(),
		SubscriptionSales:        allSales.SubscriptionSales(),
		CustomerTypeSales:        allSales.CustomerTypeSales(),
		CurrencySales:            allSales.GroupByCurrency(),
		WeekdaySales:             allSales.GroupByWeekday(),
		CustomerCount:            len(customers),
		AnnualSubscriptionCount:  len(allSales.ByAnnualSubscription().Customers()),
		MonthlySubscriptionCount: len(allSales.ByMonthlySubscription().Customers()),
		FreeSubscriptionCount:    len(allSales.ByFreeSubscription().Customers()),
	}, nil
}

func toFloat(v interface{}) (float64, error) {
	switch n := v.(type) {
	case int:
		return float64(n), nil
	case int32:
		return float64(n), nil
	case int64:
		return float64(n), nil
	case float32:
		return float64(n), nil
	case float64:
		return n, nil
	case marketplace.Amount:
		return float64(n), nil
	}
	return 0.0, fmt.Errorf("unknown type: %T", v)
}

func (r HTMLReport) Generate() (string, error) {
	funcMap := template.FuncMap{
		"addInt": func(a, b int) int {
			return a + b
		},
		"subInt": func(a, b int) int {
			return a - b
		},
		"formatInt": func(n int) string {
			return util.FormatInt(n)
		},
		"percentage": func(a, b interface{}) (string, error) {
			f1, err := toFloat(a)
			if err != nil {
				return "", err
			}

			f2, err := toFloat(b)
			if err != nil {
				return "", err
			}

			return fmt.Sprintf("%.2f%%", f1/f2*100.0), nil
		},
		"growthPercentage": func(a, b interface{}) (string, error) {
			oldValue, err := toFloat(a)
			if err != nil {
				return "", err
			}

			newValue, err := toFloat(b)
			if err != nil {
				return "", err
			}

			return fmt.Sprintf("%+.0f%%", ((newValue/oldValue)-1.0)*100.0), nil
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
