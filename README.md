# pitchfork-dl [![Build Status](https://travis-ci.org/microamp/pitchfork-dl.svg?branch=master)](https://travis-ci.org/microamp/pitchfork-dl)

Download all [Pitchfork](http://pitchfork.com/reviews/albums/) album reviews in JSON format

## Demo

[![asciicast](https://asciinema.org/a/8d9aynoywjmlkew7879pkcv81.png)](https://asciinema.org/a/8d9aynoywjmlkew7879pkcv81)

## Installation
```
go get -u github.com/microamp/pitchfork-dl
```

## Usage
```
pitchfork-dl -h
```
```
Usage of pitchfork-dl:
  -first int
    	First page (default 1)
  -last int
    	Last page
  -output string
    	Output directory (default "reviews")
  -proxy string
    	Proxy server (default "socks5://127.0.0.1:9150")
```

## Quickstart

```bash
go build pitchfork-dl.go
```

### Reviews in first 10 pages
```
pitchfork-dl -first 1 -last 10
```

### Reviews from 50th page to 100th page
```
pitchfork-dl -first 50 -last 100
```

### All reviews (as of November 2016)
```
pitchfork-dl -first 1 -last 1521
```

## License

The BSD 3-Clause License
