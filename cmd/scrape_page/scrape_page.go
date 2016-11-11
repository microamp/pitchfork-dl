package main

import (
	"flag"
	"log"
	"os"
	"strings"

	app "github.com/microamp/pitchfork-dl/app"
)

var (
	pageNumber int
	proxy      string
)

func main() {
	// Parse params
	flags := flag.NewFlagSet("pitchfork-dl-scrape-page", flag.ExitOnError)
	flags.IntVar(&pageNumber, "page", 1, "Page number")
	flags.StringVar(&proxy, "proxy", "socks5://127.0.0.1:9150", "Proxy server")
	err := flags.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}

	// Prepare proxy client
	proxyClient, err := app.GetClient(proxy)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	reviewScraper := app.Scraper{Client: proxyClient}

	requested, resp, err := reviewScraper.ScrapePage(pageNumber)
	if err != nil {
		log.Fatalf("Error while requesting %s: %s", requested, err)
	}

	reviewIDs, err := app.ParsePage(pageNumber, resp)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	log.Printf("Page %d: %d reviews found", pageNumber, len(reviewIDs))
	log.Printf("Review IDs: %s", strings.Join(reviewIDs, ", "))
}
