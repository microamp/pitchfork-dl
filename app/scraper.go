package scraper

// ReviewScraper scrapes review pages and reviews
import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"

	"golang.org/x/net/proxy"
)

var reviewBaseURL = "http://pitchfork.com/reviews/albums/"

// Scraper ...
type Scraper struct {
	Client          *http.Client
	OutputDirectory string
}

// Config ...
type Config struct {
	ProxyRawURL     string `json:"proxyRawURL"`
	OutputDirectory string `json:"outputDirectory"`
}

// ScrapePage ...
func (scraper *Scraper) ScrapePage(pageNumber int) (string, *http.Response, error) {
	requested := fmt.Sprintf("%s?page=%d", reviewBaseURL, pageNumber)
	resp, err := scraper.Client.Get(requested)
	if err != nil {
		return requested, nil, err
	}
	return requested, resp, nil
}

// ScrapeReview ...
func (scraper *Scraper) ScrapeReview(reviewID string) (string, *http.Response, error) {
	requested := fmt.Sprintf("%s%s", reviewBaseURL, reviewID)
	resp, err := scraper.Client.Get(requested)
	if err != nil {
		return requested, nil, err
	}
	return requested, resp, nil
}

// ParsePage ...
func ParsePage(pageNumber int, resp *http.Response) ([]string, error) {
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}

	page := &Page{doc: doc, PageNumber: pageNumber}

	reviewIDs := page.getReviewIDs()

	return reviewIDs, nil
}

// ParseReview ...
func ParseReview(reviewID string, resp *http.Response) (*Review, error) {
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}

	review := &Review{doc: doc, ReviewID: reviewID}

	review.setAlbums()
	review.setAuthors()
	review.setGenres()
	review.setArticle()

	return review, nil
}

// GetProxyClient gets a proxy client
func GetProxyClient(rawURL string) (*http.Client, error) {
	proxyURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{Dial: dialer.Dial}
	return &http.Client{Transport: transport}, nil
}
