package main

import (
	"fmt"
	"log"

	"github.com/jansorg/marketplace-stats/marketplace"
	"github.com/jansorg/marketplace-stats/report"
)

func fatalOpt(err error) {
	if err != nil {
		log.Fatalf("Unexpected error: %v", err)
	}
}

func main() {
	client := NewDummyClient(60, 2000, 14)

	pluginInfo, err := client.GetCurrentPluginInfo()
	fatalOpt(err)

	sales, err := client.GetAllSalesInfo()
	fatalOpt(err)

	r, err := report.NewReport(pluginInfo, sales, marketplace.Transactions{}, client, 7)
	fatalOpt(err)

	html, err := r.Generate(false)
	fatalOpt(err)
	fmt.Println(html)
}
