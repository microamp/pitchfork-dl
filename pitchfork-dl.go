package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	app "github.com/microamp/pitchfork-dl/app"
)

var (
	maxPageWorkers   = 2  // Number of concurrent page workers
	maxReviewWorkers = 48 // Number of concurrent review workers

	retryDelayPage   = 5 * time.Second
	retryDelayReview = 3 * time.Second

	proxy, output       string
	pageFirst, pageLast int
)

func writeToFile(path string, review *app.Review) error {
	bytes, err := json.MarshalIndent(review, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, bytes, 0644)
}

func startPageWorker(scraper *app.Scraper, chanPages <-chan int, chanReviewIDs chan<- string, chanDone chan<- bool, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		pageNumber, ok := <-chanPages
		if !ok {
			break // Termination due to channel closed
		}

		for {
			_, resp, err := scraper.ScrapePage(pageNumber)
			if err != nil {
				log.Fatalf("Error while scraping page, %d: %s", pageNumber, err)
			}
			if resp.StatusCode == http.StatusNotFound {
				log.Printf("Page %d not found", pageNumber)
				chanDone <- true
				break
			}

			reviewIDs, err := app.ParsePage(pageNumber, resp)
			if err != nil {
				log.Fatalf("Error while parsing page, %d: %s", pageNumber, err)
			}

			if len(reviewIDs) == 0 {
				// Retry if no data
				log.Printf("Empty data received. Retrying page %d after %d seconds...", pageNumber, retryDelayPage/time.Second)
				time.Sleep(retryDelayPage)
				continue
			} else {
				// Log page info
				go log.Printf("Page %d with %d reviews", pageNumber, len(reviewIDs))
			}

			for _, reviewID := range reviewIDs {
				chanReviewIDs <- reviewID
			}
			break
		}
	}
}

func startReviewWorker(scraper *app.Scraper, chanReviewIDs <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		reviewID, ok := <-chanReviewIDs
		if !ok {
			break // Termination due to channel closed
		}

		for {
			_, resp, err := scraper.ScrapeReview(reviewID)
			if err != nil {
				log.Fatalf("Error while scraping review, %s: %s", reviewID, err)
			}
			if resp.StatusCode == http.StatusNotFound {
				log.Printf("Review not found: %s", reviewID)
				break
			}

			review, err := app.ParseReview(reviewID, resp)
			if err != nil {
				log.Fatalf("Error while parsing review, %s: %s", reviewID, err)
			}

			if len(review.Albums) == 0 {
				// Retry if no data
				log.Printf("Empty data received. Retrying review %s after %d seconds...", reviewID, retryDelayReview/time.Second)
				time.Sleep(retryDelayReview)
				continue
			} else {
				// Log review info
				go review.PrintInfo()
			}

			// Write to file (JSON)
			filename := fmt.Sprintf("%s/%s.json", scraper.OutputDirectory, reviewID)
			err = writeToFile(filename, review)
			if err != nil {
				log.Fatalf("Error while writing to file, %s: %s", filename, err)
			}
			break
		}
	}
}

func startProcessing(scraper *app.Scraper, first, last int) {
	// Channels
	chanSigs := make(chan os.Signal)
	signal.Notify(chanSigs, os.Interrupt)
	signal.Notify(chanSigs, syscall.SIGTERM)

	chanPages := make(chan int)
	chanReviewIDs := make(chan string)

	chanDone := make(chan bool, maxPageWorkers) // Buffered!

	// Wait groups
	var wgPages, wgReviewIDs sync.WaitGroup
	wgPages.Add(maxPageWorkers)
	wgReviewIDs.Add(maxReviewWorkers)

	// Start page workers
	go func() {
		for i := 0; i < maxPageWorkers; i++ {
			go startPageWorker(scraper, chanPages, chanReviewIDs, chanDone, &wgPages)
		}
	}()

	// Start review workers
	go func() {
		for i := 0; i < maxReviewWorkers; i++ {
			go startReviewWorker(scraper, chanReviewIDs, &wgReviewIDs)
		}
	}()

	// Detect OS interrupt signals
	go func() {
		for s := range chanSigs {
			log.Printf("Shutdown signal received (%s)", s)
			chanDone <- true
			break
		}
	}()

	pageNumber := first
	for {
		select {
		case <-chanDone:
			log.Println("Waiting for workers to finish...")
			close(chanPages)
			wgPages.Wait()
			close(chanReviewIDs)
			wgReviewIDs.Wait()
			os.Exit(1)
		default:
			// If last page is unknown, scrape until the end (i.e. 404)
			if last == 0 {
				chanPages <- pageNumber
				pageNumber++
			} else {
				if pageNumber <= last {
					chanPages <- pageNumber
					pageNumber++
				} else {
					log.Println("No more pages to download")
					chanDone <- true
				}
			}
		}
	}
}

func main() {
	// Parse params
	flags := flag.NewFlagSet("pitchfork-dl", flag.ExitOnError)
	flags.StringVar(&proxy, "proxy", "socks5://127.0.0.1:9150", "Proxy server")
	flags.StringVar(&output, "output", "reviews", "Output directory")
	flags.IntVar(&pageFirst, "first", 1, "First page")
	flags.IntVar(&pageLast, "last", 0, "Last page")

	err := flags.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}

	if pageLast == 0 {
		log.Printf("Scraping pages from %d...", pageFirst)
	} else {
		log.Printf("Scraping pages from %d to %d...", pageFirst, pageLast)
	}

	// Prepare proxy client
	proxyClient, err := app.GetProxyClient(proxy)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Fire up goroutines and start processing
	scraper := &app.Scraper{
		Client:          proxyClient,
		OutputDirectory: output,
	}
	startProcessing(scraper, pageFirst, pageLast)
}
