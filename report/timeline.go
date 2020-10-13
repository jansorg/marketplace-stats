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

type TimelineDay struct {
	Day       marketplace.YearMonthDay
	Downloads int
	Sales     marketplace.Sales
}

type Timeline struct {
	Days []TimelineDay
}

func NewTimeline(sales marketplace.Sales, downloads []marketplace.DownloadsDaily) *Timeline {
	var firstDay, lastDay marketplace.YearMonthDay
	if len(sales) > 0 {
		firstDay = sales[0].Date
		lastDay = sales[len(sales)-1].Date
	}

	if len(downloads) > 0 {
		if first := downloads[0].Date(); firstDay.IsAfter(first) {
			firstDay = first
		}

		if last := downloads[len(downloads)-1].Date(); last.IsAfter(lastDay) {
			lastDay = last
		}
	}

	var days []TimelineDay
	day := firstDay
	for !day.IsAfter(lastDay) {
		var downloadCount int
		for _, d := range downloads {
			if d.Date().Equals(day) {
				downloadCount = d.Downloads
				break
			}
		}

		days = append(days, TimelineDay{
			Day:       day,
			Downloads: downloadCount,
			Sales:     sales.ByYearMonthDay(day),
		})

		day = day.NextDay()
	}

	return &Timeline{
		Days: days,
	}
}

func (t *Timeline) DrawSVG() template.HTML {
	var xValues []time.Time
	var downloadValues []float64
	var maxDownload float64
	var salesUSDValues []float64
	var maxSale float64
	for _, d := range t.Days {
		xValues = append(xValues, d.Day.AsDate())
		downloadValues = append(downloadValues, float64(d.Downloads))
		usdValue := float64(d.Sales.TotalSumUSD())
		salesUSDValues = append(salesUSDValues, usdValue)

		maxDownload = math.Max(maxDownload, float64(d.Downloads))
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

	graph := chart.Chart{
		XAxis: chart.XAxis{
			Name: "Downloads",
		},
		YAxis: chart.YAxis{
			Name: "Sales in USD",
		},
		Series: []chart.Series{
			salesSeries,
		},
	}

	out := strings.Builder{}
	err := graph.Render(canvas.NewGoChart(svg.Writer), &out)
	if err != nil {
		panic(err)
	}

	return template.HTML(out.String())
}
