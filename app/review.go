package scraper

import (
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Review ...
type Review struct {
	doc      *goquery.Document
	ReviewID string   `json:"id"`
	Albums   []*Album `json:"albums"`
	Authors  []string `json:"authors"`
	Genres   []string `json:"genres"`
	Article  string
}

// Album ...
type Album struct {
	selection *goquery.Selection
	Artists   []string `json:"artists"`
	Title     string   `json:"title"`
	Score     string   `json:"score"`
	Labels    []string `json:"labels"`
	Year      string   `json:"year"`
}

func (album *Album) setArtists() {
	artists := []string{}
	album.selection.Find("h2.artists ul.artist-list").Each(
		func(i int, s *goquery.Selection) {
			s.Find("li a").Each(
				func(i int, s *goquery.Selection) {
					artists = append(artists, s.Text())
				},
			)
		},
	)
	album.Artists = artists
}

func (album *Album) setTitle() {
	title := album.selection.Find("h1.review-title").Text()
	album.Title = title
}

func (album *Album) setScore() {
	score := album.selection.Find("div.score-box div.score-circle span.score").Text()
	album.Score = score
}

func (album *Album) setLabels() {
	labels := []string{}
	album.selection.Find("div.album-art div.labels-and-years ul.label-list").Each(
		func(i int, s *goquery.Selection) {
			s.Find("li").Each(
				func(i int, s *goquery.Selection) {
					labels = append(labels, s.Text())
				},
			)
		},
	)
	album.Labels = labels
}

func (album *Album) setYear() {
	year := album.selection.Find("div.album-art div.labels-and-years span.year").Text()
	album.Year = strings.TrimLeft(year, "• ")
}

func (review *Review) setAlbums() {
	albums := []*Album{}

	review.doc.Find("div.review-detail article").Each(
		func(_ int, s *goquery.Selection) {
			s.Find("div.tombstone div.row").Each(
				func(i int, s *goquery.Selection) {
					album := &Album{selection: s}

					album.setArtists()
					album.setTitle()
					album.setScore()
					album.setLabels()
					album.setYear()

					albums = append(albums, album)
				},
			)
		},
	)

	review.Albums = albums
}

func (review *Review) setAuthors() {
	authors := []string{}
	review.doc.Find("div.review-body div.article-meta ul.authors-detail").Each(
		func(_ int, s *goquery.Selection) {
			s.Find("li div a").Each(
				func(_ int, s *goquery.Selection) {
					authors = append(authors, s.Text())
				},
			)
		},
	)
	review.Authors = authors
}

func (review *Review) setGenres() {
	genres := []string{}
	review.doc.Find("ul.genre-list").Each(
		func(_ int, s *goquery.Selection) {
			s.Find("li a").Each(
				func(_ int, s *goquery.Selection) {
					genres = append(genres, s.Text())
				},
			)
		},
	)
	review.Genres = genres
}

func (review *Review) setArticle() {
	paragraphs := []string{}
	review.doc.Find("div.review-body div.article-content div.review-text div.contents").Each(
		func(_ int, s *goquery.Selection) {
			s.Find("p").Each(
				func(_ int, s *goquery.Selection) {
					paragraphs = append(paragraphs, s.Text())
				},
			)
		},
	)
	review.Article = strings.Join(paragraphs, "\n\n")
}

// PrintInfo ...
func (review *Review) PrintInfo() {
	for _, album := range review.Albums {
		log.Printf(
			"Review ID: %s | Artists: %s | Title: %s | Score: %s | Year: %s | Authors: %s | Genres: %s",
			review.ReviewID,
			strings.Join(album.Artists, "/"),
			album.Title,
			album.Score,
			album.Year,
			strings.Join(review.Authors, "/"),
			strings.Join(review.Genres, "/"),
		)
	}
}
