package main

import (
	"github.com/jansorg/go-randomdata"
	"github.com/jansorg/marketplace-stats/marketplace"
	"math/rand"
	"strconv"
	"time"
)

var currencies = []marketplace.Currency{"USD", "EUR", "CZK"}

func randomCustomers(max int) marketplace.Customers {
	rand.Seed(time.Now().UnixNano())

	var result marketplace.Customers
	for i := 0; i < max; i++ {
		account := marketplace.AccountTypePersonal
		if rand.Intn(100)%3 == 0 {
			account = marketplace.AccountTypeOrganization
		}

		var name string
		if account == marketplace.AccountTypeOrganization {
			name = randomdata.SillyName()
		} else {
			name = randomdata.FullName(randomdata.RandomGender)
		}

		result = append(result, marketplace.NewCustomer(
			marketplace.CustomerID(i),
			name,
			randomdata.Country(randomdata.FullCountry),
			account,
		))
	}
	return result
}

func randomSales(max int, maxMonths int, customers marketplace.Customers) marketplace.Sales {
	rand.Seed(time.Now().UnixNano())

	now := time.Now()

	var result marketplace.Sales
	for i := 0; i < max; i++ {
		refID := strconv.Itoa(i)
		saleDate := now.AddDate(0, -rand.Intn(maxMonths), 14-int(rand.Float64()*28))

		subscription := marketplace.AnnualSubscription
		if rand.Intn(100)%3 == 0 {
			subscription = marketplace.MonthlySubscription
		}

		customer := customers[rand.Intn(len(customers)-1)]
		amount := rand.Float64() * 30
		amountUSD := amount * (1 + rand.Float64())

		if customer.Type == marketplace.AccountTypeOrganization {
			amount *= 3
			amountUSD *= 3
		}

		result = append(result, marketplace.NewSale(
			refID,
			saleDate.Year(),
			int(saleDate.Month()),
			saleDate.Day(),
			subscription,
			customer,
			marketplace.Amount(amount),
			currencies[rand.Intn(len(currencies)-1)],
			marketplace.Amount(amountUSD),
		))
	}

	return result
}

func randomDownloads(sales marketplace.Sales) []marketplace.DownloadAndDate {
	now := time.Now()
	rand.Seed(now.UnixNano())

	var result []marketplace.DownloadAndDate
	for _, sale := range sales {
		result = append(result, marketplace.DownloadAndDate{
			rand.Intn(1000),
			sale.Date.Year(),
			sale.Date.Month(),
			sale.Date.Day(),
		})
	}
	return result
}
