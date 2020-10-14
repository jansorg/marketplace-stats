package report

import (
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/svg"
	"github.com/wcharczuk/go-chart"
	"html/template"
	"math"
	"strings"
	"time"

	"github.com/jansorg/marketplace-stats/marketplace"
)

type AsDate interface {
	AsDate() time.Time
}

type TimelineItem struct {
	Date            AsDate
	Downloads       int
	DownloadsUnique int
	Sales           marketplace.Sales
}

type Timeline struct {
	Items []TimelineItem
}

func NewMonthlyTimeline(sales marketplace.Sales, downloads []marketplace.DownloadMonthly) *Timeline {
	var firstMonth, lastMonth marketplace.YearMonth
	if len(sales) > 0 {
		firstMonth = sales[0].Date.AsYearMonth()
		lastMonth = sales[len(sales)-1].Date.AsYearMonth()
	}

	if len(downloads) > 0 {
		if first := downloads[0].Date(); firstMonth.IsAfter(first) {
			firstMonth = first
		}

		if last := downloads[len(downloads)-1].Date(); last.IsAfter(lastMonth) {
			lastMonth = last
		}
	}

	month := firstMonth

	var months []TimelineItem
	for !month.IsAfter(lastMonth) {
		var downloadCount int
		for _, d := range downloads {
			if d.Date().Equals(month) {
				downloadCount = d.Downloads
				break
			}
		}

		var downloadUniqueCount int
		for _, d := range downloads {
			if d.Date().Equals(month) {
				downloadUniqueCount = d.Downloads
				break
			}
		}

		months = append(months, TimelineItem{
			Date:            month,
			Downloads:       downloadCount,
			DownloadsUnique: downloadUniqueCount,
			Sales:           sales.ByYearMonth(month),
		})

		month = month.NextMonth()
	}

	return &Timeline{
		Items: months,
	}
}

func NewWeeklyTimeline(sales marketplace.Sales, downloads []marketplace.DownloadAndDate) *Timeline {
	var firstMonth, lastMonth marketplace.YearMonthDay
	if len(sales) > 0 {
		firstMonth = sales[0].Date
		lastMonth = sales[len(sales)-1].Date
	}

	if len(downloads) > 0 {
		if first := downloads[0].Date(); firstMonth.IsAfter(first) {
			firstMonth = first
		}

		if last := downloads[len(downloads)-1].Date(); last.IsAfter(lastMonth) {
			lastMonth = last
		}
	}

	week := firstMonth

	var months []TimelineItem
	for !week.IsAfter(lastMonth) {
		var downloadCount int
		for _, d := range downloads {
			if d.Date().Equals(week) {
				downloadCount = d.Downloads
				break
			}
		}

		var downloadUniqueCount int
		for _, d := range downloads {
			if d.Date().Equals(week) {
				downloadUniqueCount = d.Downloads
				break
			}
		}

		nextWeek := week.AddDays(7)
		months = append(months, TimelineItem{
			Date:            week,
			Downloads:       downloadCount,
			DownloadsUnique: downloadUniqueCount,
			Sales:           sales.ByDateRange(week, nextWeek),
		})

		week = nextWeek
	}

	return &Timeline{
		Items: months,
	}
}

func NewDailyTimeline(sales marketplace.Sales, daily []marketplace.DownloadAndDate) *Timeline {
	var firstDay, lastDay marketplace.YearMonthDay
	if len(sales) > 0 {
		firstDay = sales[0].Date
		lastDay = sales[len(sales)-1].Date
	}

	if len(daily) > 0 {
		if first := daily[0].Date(); firstDay.IsAfter(first) {
			firstDay = first
		}

		if last := daily[len(daily)-1].Date(); last.IsAfter(lastDay) {
			lastDay = last
		}
	}

	var days []TimelineItem
	day := firstDay
	for !day.IsAfter(lastDay) {
		var downloadCount int
		for _, d := range daily {
			if d.Date().Equals(day) {
				downloadCount = d.Downloads
				break
			}
		}

		days = append(days, TimelineItem{
			Date:      day,
			Downloads: downloadCount,
			Sales:     sales.ByYearMonthDay(day),
		})

		day = day.NextDay()
	}

	return &Timeline{
		Items: days,
	}
}

func (t *Timeline) DrawSVG() template.HTML {
	var xValues []time.Time
	var xAxisTicks []chart.Tick

	var downloadValues []float64
	var downloadUniqueValues []float64
	var maxDownload float64
	var salesUSDValues []float64
	var maxSale float64
	for _, d := range t.Items {
		xValues = append(xValues, d.Date.AsDate())

		downloadValues = append(downloadValues, float64(d.Downloads))
		downloadUniqueValues = append(downloadUniqueValues, float64(d.DownloadsUnique))
		usdValue := float64(d.Sales.TotalSumUSD())
		salesUSDValues = append(salesUSDValues, usdValue)

		maxDownload = math.Max(maxDownload, math.Max(float64(d.Downloads), float64(d.DownloadsUnique)))
		maxSale = math.Max(maxSale, usdValue)
	}

	salesSeries := chart.TimeSeries{
		Style: chart.Style{
			StrokeColor: chart.ColorBlue,
			FillColor:   chart.ColorBlue.WithAlpha(100),
			ClassName:   "line-sales",
		},
		XValues: xValues,
		YValues: salesUSDValues,
	}

	downloadsUniqueSeries := chart.TimeSeries{
		Style: chart.Style{
			StrokeColor: chart.ColorBlack,
			ClassName:   "line-download-unique",
		},
		XValues: xValues,
		YValues: downloadUniqueValues,
		YAxis:   chart.YAxisSecondary,
	}

	graph := chart.Chart{
		XAxis: chart.XAxis{
			TickPosition: chart.TickPositionUnderTick,
			Ticks:        xAxisTicks,
		},
		YAxis: chart.YAxis{
			ValueFormatter: func(v interface{}) string {
				return chart.FloatValueFormatterWithFormat(v, "%.0f USD")
			},
		},
		YAxisSecondary: chart.YAxis{
			ValueFormatter: chart.IntValueFormatter,
		},
		Series: []chart.Series{
			salesSeries,
			downloadsUniqueSeries,
		},
	}

	out := strings.Builder{}
	err := graph.Render(canvas.NewGoChart(svg.Writer), &out)
	if err != nil {
		panic(err)
	}

	return template.HTML(out.String())
}
