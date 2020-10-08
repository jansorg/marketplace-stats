package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"jansorg/marketplace-stats/marketplace"
	"jansorg/marketplace-stats/report"
	"jansorg/marketplace-stats/statistic"
)

func fatalOpt(err error) {
	if err != nil {
		log.Fatalf("Unexpected error: %v", err)
	}
}

func main() {
	pluginIDFlag := flag.String("pluginID", "13841", "The ID of the plugin, e.g. 12345.")
	fetchOnlineFlag := flag.Bool("fetch", false, "If online data should be fetched. Needs the --token flag.")
	tokenParam := flag.String("token", "", "The token to access the API of the JetBrains marketplace.")
	tokenFileParam := flag.String("tokenFile", "", "Path to a file, which contains the token")
	fileParam := flag.String("cache-file", "sales.json", "The file where sales data is cached. Use -fetch to update it.")
	reportFileParam := flag.String("html", "report.html", "The file where the HTML sales report is saved.")
	flag.Parse()

	if *fetchOnlineFlag && *tokenParam == "" && *tokenFileParam == "" {
		fmt.Fprintf(os.Stderr, "Unable to load sales data without a token. Please provide the marketplace API token.\n")
		return
	}

	var sales []marketplace.Sale
	var pluginInfo marketplace.PluginInfo

	token, err := getToken(*tokenParam, *tokenFileParam)
	fatalOpt(err)
	client := marketplace.NewClient(*pluginIDFlag, token)

	pluginInfo, err = client.GetCurrentPluginInfo()
	fatalOpt(err)

	monthlyDownloadsUnique, err := client.DownloadsMonthly(false, "", "", "", "", "")
	fatalOpt(err)
	monthlyDownloadsTotal, err := client.DownloadsMonthly(true, "", "", "", "", "")
	fatalOpt(err)

	if *fetchOnlineFlag {
		sales, err = client.GetAllSalesInfo()
		fatalOpt(err)

		// write to cache file
		cacheFile, err := os.OpenFile(*fileParam, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		defer cacheFile.Close()
		encoder := json.NewEncoder(cacheFile)
		encoder.SetIndent("", "  ")
		err = encoder.Encode(sales)
		fatalOpt(err)
	} else {
		file, err := os.Open(*fileParam)
		fatalOpt(err)

		err = json.NewDecoder(file).Decode(&sales)
		fatalOpt(err)
	}

	// iterate months
	var months []*statistic.Month
	if len(sales) > 0 {
		firstDate := sales[0].Date.AsDate()
		lastDate := sales[len(sales)-1].Date.AsDate().AddDate(0, 1, 0)
		yearMonth := firstDate

		now := time.Now().In(marketplace.ServerTimeZone)
		lastDateCurrentMonth := now.AddDate(0, 1, -now.Day())

		var prevMonthData *statistic.Month
		for !yearMonth.After(lastDate) && !yearMonth.After(lastDateCurrentMonth) {
			month := statistic.NewMonthForDate(yearMonth)
			month.Update(sales, prevMonthData, monthlyDownloadsTotal, monthlyDownloadsUnique)

			months = append(months, month)
			yearMonth = yearMonth.AddDate(0, 1, 0)
			prevMonthData = month
		}
	}

	// iterate years
	var years []*statistic.Year
	if len(sales) > 0 {
		firstDate := sales[0].Date.AsDate()
		lastDate := sales[len(sales)-1].Date.AsDate().AddDate(0, 1, 0)
		year := firstDate

		for !year.After(lastDate) {
			stats := statistic.NewYear(year.Year())
			stats.Update(sales)

			years = append(years, stats)
			year = year.AddDate(1, 0, 0)
		}
	}

	if *reportFileParam != "" {
		r := report.NewReport(pluginInfo, sales, years, months)
		html, err := r.Generate()
		fatalOpt(err)

		err = ioutil.WriteFile(*reportFileParam, []byte(html), 0600)
		fatalOpt(err)
	}
}

func getToken(token string, tokenFile string) (string, error) {
	if token != "" {
		return token, nil
	}
	if tokenFile != "" {
		data, err := ioutil.ReadFile(tokenFile)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	return "", fmt.Errorf("missing token")
}
