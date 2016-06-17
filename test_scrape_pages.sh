#!/usr/bin/env bash

Echo "Testing: page 1"
go run cmd/scrape_page/scrape_page.go

Echo "Testing: page 100"
go run cmd/scrape_page/scrape_page.go -page 100
