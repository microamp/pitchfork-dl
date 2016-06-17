#!/usr/bin/env bash

echo "Testing: single artist, single album"
go run cmd/scrape_review/scrape_review.go -id 21715-nocturnal-koreans

echo "Testing: single artist, multiple albums"
go run cmd/scrape_review/scrape_review.go -id 21950-elseq-1-5

echo "Testing: multiple artists (collaboration), multiple albums"
go run cmd/scrape_review/scrape_review.go -id 21821-everythings-beautiful-miles-ahead-ost
