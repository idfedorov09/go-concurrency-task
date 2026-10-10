package main

import (
	"flag"
	"fmt"
	"testing"
	"time"
	"github.com/idfedorov09/go-concurrency-task/internal/crawl"
)

var depthPtr = flag.Int("depth", 2, "depth of crawl")
var sleepMsPtr = flag.Int("fetch-sleep", 0, "sleep imitation on fetch")

type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	// fetch work imitation, default off
	if *sleepMsPtr > 0 {
		time.Sleep(time.Duration(*sleepMsPtr)*time.Millisecond)
	}

	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}

func TestCrawl(t *testing.T) {
	CrawlElapsed("https://golang.org/", *depthPtr, fetcher)
}

func CrawlElapsed(targetUrl string, depth int, fetcher crawl.Fetcher) {
	start := time.Now()
	Crawl(targetUrl, depth, fetcher)
	elapsed := time.Since(start)

	fmt.Printf("\n----- Crawl execution time: %8d ns -----\n\n", elapsed.Nanoseconds())
}