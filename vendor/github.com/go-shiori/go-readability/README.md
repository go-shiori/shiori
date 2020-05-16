Go-Readability
===

[![GoDoc](https://godoc.org/github.com/go-shiori/go-readability?status.png)](https://godoc.org/github.com/go-shiori/go-readability)
[![Travis CI](https://travis-ci.org/go-shiori/go-readability.svg?branch=master)](https://travis-ci.org/go-shiori/go-readability)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-shiori/go-readability)](https://goreportcard.com/report/github.com/go-shiori/go-readability)
[![Donate PayPal](https://img.shields.io/static/v1?label=donate&message=PayPal&color=00457C&logo=paypal)](https://www.paypal.me/RadhiFadlillah)
[![Donate Ko-fi](https://img.shields.io/static/v1?label=donate&message=Ko-fi&color=F16061&logo=ko-fi)](https://ko-fi.com/radhifadlillah)

Go-Readability is a Go package that find the main readable content and the metadata from a HTML page. It works by removing clutter like buttons, ads, background images, script, etc.

This package is based from [Readability.js](https://github.com/mozilla/readability) by [Mozilla](https://github.com/mozilla), and written line by line to make sure it looks and works as similar as possible. This way, hopefully all web page that can be parsed by Readability.js are parse-able by go-readability as well.

## Status

This package is stable enough for use and up to date with Readability.js until commit [`d5621f8`](https://github.com/mozilla/readability/commit/d5621f85e775229332bf0f6f2b1d3d789c638f2d).

## Installation

To install this package, just run `go get` :

```
go get -u -v github.com/go-shiori/go-readability
```

## Example

To get the readable content from an URL, you can use `readability.FromURL`. It will fetch the web page from specified url, check if it's readable, then parses the response to find the readable content :

```go
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	readability "github.com/go-shiori/go-readability"
)

var (
	urls = []string{
		// this one is article, so it's parse-able
		"https://www.nytimes.com/2019/02/20/climate/climate-national-security-threat.html",
		// while this one is not an article, so readability will fail to parse.
		"https://www.nytimes.com/",
	}
)

func main() {
	for i, url := range urls {
		article, err := readability.FromURL(url, 30*time.Second)
		if err != nil {
			log.Fatalf("failed to parse %s, %v\n", url, err)
		}

		dstTxtFile, _ := os.Create(fmt.Sprintf("text-%02d.txt", i+1))
		defer dstTxtFile.Close()
		dstTxtFile.WriteString(article.TextContent)

		dstHTMLFile, _ := os.Create(fmt.Sprintf("html-%02d.html", i+1))
		defer dstHTMLFile.Close()
		dstHTMLFile.WriteString(article.Content)

		fmt.Printf("URL     : %s\n", url)
		fmt.Printf("Title   : %s\n", article.Title)
		fmt.Printf("Author  : %s\n", article.Byline)
		fmt.Printf("Length  : %d\n", article.Length)
		fmt.Printf("Excerpt : %s\n", article.Excerpt)
		fmt.Printf("SiteName: %s\n", article.SiteName)
		fmt.Printf("Image   : %s\n", article.Image)
		fmt.Printf("Favicon : %s\n", article.Favicon)
		fmt.Printf("Text content saved to \"text-%02d.txt\"\n", i+1)
		fmt.Printf("HTML content saved to \"html-%02d.html\"\n", i+1)
		fmt.Println()
	}
}
```

However, sometimes you want to parse an URL no matter if it's an article or not. For example is when you only want to get metadata of the page. To do that, you have to download the page manually using `http.Get`, then parse it using `readability.FromReader` :

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	readability "github.com/go-shiori/go-readability"
)

var (
	urls = []string{
		// Both will be parse-able now
		"https://www.nytimes.com/2019/02/20/climate/climate-national-security-threat.html",
		// But this one will not have any content
		"https://www.nytimes.com/",
	}
)

func main() {
	for _, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			log.Fatalf("failed to download %s: %v\n", url, err)
		}
		defer resp.Body.Close()

		article, err := readability.FromReader(resp.Body, url)
		if err != nil {
			log.Fatalf("failed to parse %s: %v\n", url, err)
		}

		fmt.Printf("URL     : %s\n", url)
		fmt.Printf("Title   : %s\n", article.Title)
		fmt.Printf("Author  : %s\n", article.Byline)
		fmt.Printf("Length  : %d\n", article.Length)
		fmt.Printf("Excerpt : %s\n", article.Excerpt)
		fmt.Printf("SiteName: %s\n", article.SiteName)
		fmt.Printf("Image   : %s\n", article.Image)
		fmt.Printf("Favicon : %s\n", article.Favicon)
		fmt.Println()
	}
}
```

## Command Line Usage

You can also use `go-readability` as command line app. To do that, first install the CLI :

```
go get -u -v github.com/go-shiori/go-readability/cmd/...
```

Now you can use it by running `go-readability` in your terminal :

```
$ go-readability -h

go-readability is parser to fetch the readable content of a web page.
The source can be an url or existing file in your storage.

Usage:
  go-readability [flags] source

Flags:
  -h, --help       help for go-readability
  -m, --metadata   only print the page's metadata
```

## Licenses

Go-Readability is distributed under [MIT license](https://choosealicense.com/licenses/mit/), which means you can use and modify it however you want. However, if you make an enhancement for it, if possible, please send a pull request. If you like this project, please consider donating to me either via [PayPal](https://www.paypal.me/RadhiFadlillah) or [Ko-Fi](https://ko-fi.com/radhifadlillah).
