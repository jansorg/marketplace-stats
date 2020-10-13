package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/jansorg/marketplace-stats/marketplace"
	"github.com/jansorg/marketplace-stats/report"
)

func fatalOpt(err error) {
	if err != nil {
		log.Fatalf("Unexpected error: %v", err)
	}
}

func main() {
	pluginID := flag.String("plugin-id", "", "The ID of the plugin, e.g. 12345.")
	tokenParam := flag.String("token", "", "The token to access the API of the JetBrains marketplace. --token-file is an alternative.")
	tokenFileParam := flag.String("token-file", "", "Path to a file, which contains the token to access the API of the JetBrains marketplace.")
	fileParam := flag.String("cache-file", "", "The file where sales data is cached. Use -fetch to update it.")
	fetchParam := flag.Bool("fetch", true, "The file where sales data is cached. Use -fetch to update it.")
	reportFileParam := flag.String("out", "report.html", "The file where the HTML sales report is saved.")
	flag.Parse()

	if *pluginID == "" {
		fmt.Fprintln(os.Stderr, "Plugin ID not defined. Use --plugin-id to define it.")
		return
	}

	if *fetchParam && *tokenParam == "" && *tokenFileParam == "" {
		fmt.Fprintln(os.Stderr, "Unable to load sales data without a token. Please provide the marketplace API token.")
		return
	}

	token, err := getToken(*tokenParam, *tokenFileParam)
	fatalOpt(err)

	var sales []marketplace.Sale
	var pluginInfo marketplace.PluginInfo

	client := marketplace.NewClient(*pluginID, token)
	pluginInfo, err = client.GetCurrentPluginInfo()
	fatalOpt(err)

	if *fetchParam {
		sales, err = client.GetAllSalesInfo()
		fatalOpt(err)

		// write to cache file, if it's defined
		if *fileParam != "" {
			cacheFile, err := os.OpenFile(*fileParam, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
			defer cacheFile.Close()
			encoder := json.NewEncoder(cacheFile)
			encoder.SetIndent("", "  ")
			err = encoder.Encode(sales)
			fatalOpt(err)
		}
	} else {
		file, err := os.Open(*fileParam)
		fatalOpt(err)
		err = json.NewDecoder(file).Decode(&sales)
		fatalOpt(err)
	}

	report, err := report.NewReport(pluginInfo, sales, client)
	fatalOpt(err)

	html, err := report.Generate()
	fatalOpt(err)

	if *reportFileParam != "" && *reportFileParam != "-" {
		err = ioutil.WriteFile(*reportFileParam, []byte(html), 0600)
		fatalOpt(err)
	} else {
		// print to stdout
		fmt.Println(html)
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
