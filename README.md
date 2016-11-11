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

### Reviews in first 10 pages
```
pitchfork-dl -last 10
```

### Reviews from 50th page to 100th page
```
pitchfork-dl -first 50 -last 100
```

### Reviews from 1000th page till the last page (unknown)
```
pitchfork-dl -first 1000
```

### All reviews
```
pitchfork-dl
```

## License

The BSD 3-Clause License
