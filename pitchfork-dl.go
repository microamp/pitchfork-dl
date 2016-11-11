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
	maxPageWorkers   = 3  // Number of concurrent page workers
	maxReviewWorkers = 12 // Number of concurrent review workers

	retryDelayPage   = 30 * time.Second
	retryDelayReview = 30 * time.Second

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

func startPageWorker(scraper *app.Scraper, chanPages <-chan int, chanDone <-chan bool, chanReviewIDs chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-chanDone:
			return
		case pageNumber := <-chanPages:
			_, resp, err := scraper.ScrapePage(pageNumber)
			if err != nil {
				log.Printf("Error while scraping page %d: %+v", pageNumber, err)
				continue
			}
			if resp.StatusCode != http.StatusOK {
				log.Printf("Status %d received", resp.StatusCode)
				continue
			}

			reviewIDs, err := app.ParsePage(pageNumber, resp)
			if err != nil {
				log.Printf("Error while parsing page %d: %+v", pageNumber, err)
				continue
			}

			noOfReviewIDs := len(reviewIDs)
			if noOfReviewIDs == 0 {
				log.Printf("Empty data received. Retrying page %d after %d seconds...", pageNumber, retryDelayPage/time.Second)
				time.Sleep(retryDelayPage)
				continue
			}

			// Log page info
			log.Printf("Page %d with %d reviews", pageNumber, noOfReviewIDs)

			for _, reviewID := range reviewIDs {
				chanReviewIDs <- reviewID
			}
		}
	}
}

func startReviewWorker(scraper *app.Scraper, chanReviewIDs <-chan string, chanDone <-chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-chanDone:
			return
		case reviewID := <-chanReviewIDs:
			_, resp, err := scraper.ScrapeReview(reviewID)
			if err != nil {
				log.Printf("Error while scraping review %s: %+v", reviewID, err)
				continue
			}
			if resp.StatusCode != http.StatusOK {
				log.Printf("Status %d received", resp.StatusCode)
				continue
			}

			review, err := app.ParseReview(reviewID, resp)
			if err != nil {
				log.Printf("Error while parsing review %s: %+v", reviewID, err)
				continue
			}

			if len(review.Albums) == 0 {
				log.Printf("Empty data received. Retrying review %s after %d seconds...", reviewID, retryDelayReview/time.Second)
				time.Sleep(retryDelayReview)
				continue
			}

			// Log review info
			review.PrintInfo()

			// Write to file as JSON
			filename := fmt.Sprintf("%s/%s.json", scraper.OutputDirectory, reviewID)
			err = writeToFile(filename, review)
			if err != nil {
				log.Printf("Error while writing to file %s: %+v", filename, err)
			}
		}
	}
}

func startProcessing(scraper *app.Scraper, first, last int) {
	// Channels
	chanSigs := make(chan os.Signal)
	signal.Notify(chanSigs, os.Interrupt)
	signal.Notify(chanSigs, syscall.SIGTERM)

	chanPages := make(chan int)
	chanReviewIDs := make(chan string, 24) // Buffered!

	chanDone1 := make(chan bool)
	chanDone2 := make(chan bool)

	// Wait groups
	var wgPages, wgReviewIDs sync.WaitGroup
	wgPages.Add(maxPageWorkers)
	wgReviewIDs.Add(maxReviewWorkers)

	// Start page workers
	go func() {
		for i := 0; i < maxPageWorkers; i++ {
			go startPageWorker(scraper, chanPages, chanDone1, chanReviewIDs, &wgPages)
		}
	}()

	// Start review workers
	go func() {
		for i := 0; i < maxReviewWorkers; i++ {
			go startReviewWorker(scraper, chanReviewIDs, chanDone2, &wgReviewIDs)
		}
	}()

	// Create page jobs
	go func() {
		for pageNumber := first; pageNumber < last; pageNumber++ {
			chanPages <- pageNumber
		}
	}()

	// Receive OS error signal
	s := <-chanSigs
	log.Printf("Signal received (%s). Broadcasting cancellation to all workers...", s)

	// Close channel to broadcast done signals to all page worker goroutines
	close(chanDone1)
	log.Printf("Waiting for all page workers to complete...")
	wgPages.Wait()

	// Close channel to broadcast done signals to all review worker goroutines
	close(chanDone2)
	log.Printf("Waiting for all review workers to complete...")
	wgReviewIDs.Wait()

	log.Printf("Exiting...")
}

func main() {
	// Parse params
	flags := flag.NewFlagSet("pitchfork-dl", flag.ExitOnError)
	flags.StringVar(&proxy, "proxy", "socks5://127.0.0.1:9150", "Proxy server")
	flags.StringVar(&output, "output", "reviews", "Output directory")
	flags.IntVar(&pageFirst, "first", 1, "First page")
	flags.IntVar(&pageLast, "last", 10, "Last page")

	err := flags.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}

	log.Printf("Scraping pages %d to %d...", pageFirst, pageLast)

	// Prepare client
	client, err := app.GetClient(proxy)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Fire up goroutines and start processing
	scraper := &app.Scraper{
		Client:          client,
		OutputDirectory: output,
	}
	startProcessing(scraper, pageFirst, pageLast)
}
