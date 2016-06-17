package scraper

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func getReviewID(albumLink string) string {
	withoutPrefix := strings.TrimLeft(albumLink, "/reviews/albums/")
	return strings.TrimRight(withoutPrefix, "/")
}

// Page ...
type Page struct {
	doc        *goquery.Document
	PageNumber int
}

func (page *Page) getReviewIDs() []string {
	reviewIDs := []string{}
	page.doc.Find("div.fragment-list").Each(
		func(_ int, s1 *goquery.Selection) {
			s1.Find("div.review").Each(
				func(_ int, s2 *goquery.Selection) {
					if path, found := s2.Find("a").Attr("href"); found {
						reviewIDs = append(reviewIDs, getReviewID(path))
					}
				},
			)
		},
	)
	return reviewIDs
}
