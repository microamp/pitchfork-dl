package main

import (
	"flag"
	"log"
	"os"

	app "github.com/microamp/pitchfork-dl/app"
)

var (
	reviewID    string
	proxyRawURL = "socks5://127.0.0.1:9150"
)

func main() {
	// Parse optional param, 'id' (i.e. review ID)
	flags := flag.NewFlagSet("pitchfork-dl-scrape-review", flag.ExitOnError)
	flags.StringVar(&reviewID, "id", "", "Review ID")
	err := flags.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}
	if reviewID == "" {
		log.Fatal("Error: review ID must be provided")
	}

	// Prepare proxy client
	proxyClient, err := app.GetProxyClient(proxyRawURL)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	reviewScraper := app.Scraper{Client: proxyClient}
	requested, resp, err := reviewScraper.ScrapeReview(reviewID)
	if err != nil {
		log.Fatalf("Error while requesting %s: %s", requested, err)
	}

	review, err := app.ParseReview(reviewID, resp)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	review.PrintInfo()
}
