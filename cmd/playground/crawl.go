package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"github.com/idfedorov09/go-concurrency-task/internal/crawl"
)

func Run() {
	fmt.Println("Используйте test для запуска тестирования функционала")
}

type Crawler struct {
	fetcher crawl.Fetcher
	cache crawl.Cache
}

func (c *Crawler) FetchPage(targetUrl string) (crawl.PageInfo, error) {
	fmt.Println(" >>> REAL FETCH:", targetUrl)
    return crawl.FetchPage(c.fetcher, targetUrl)
}

func WithLogging(deps *Crawler, next crawl.FetchFunc) crawl.FetchFunc {
	return func(targetUrl string) (crawl.PageInfo, error) {
		page, err := next(targetUrl)

		if err != nil {
			fmt.Println(err)
			return crawl.PageInfo{}, nil
		}

		fmt.Printf("found: %s %q\n", targetUrl, page.Body)
		return page, nil
	}
}

func WithCache(deps *Crawler, next crawl.FetchFunc) crawl.FetchFunc {
	return func(targetUrl string) (crawl.PageInfo, error) {
		if p, err, ok := deps.cache.Has(targetUrl); ok {
			return p, err
		}
		p, err, _ := deps.cache.PutResult(targetUrl, next)
		return p, err
	}
}

func Crawl(startUrl string, depth int, fetcher crawl.Fetcher) {
	c := Crawler{
		fetcher: fetcher,
		cache: crawl.MakeCache(),
	}

	session := crawl.NewCrawlSession(&c).
						Wrap(WithCache).
						Wrap(WithLogging)
	
	type depthedType struct {
		targetUrl string
		depth int
	}

	var pending atomic.Int32
	pending.Store(1)
	queue := make(chan depthedType, 1)
	queue <- depthedType{startUrl, depth}

	wg := sync.WaitGroup{}

	for cur := range queue {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if pending.Add(-1) == 0 {
					close(queue)
				}
			}()

			urlParsed := crawl.CrawlPage(cur.targetUrl, session)

			for nextUrl := range urlParsed {
				if cur.depth > 0 {
					pending.Add(1)
					queue <- depthedType{nextUrl, cur.depth - 1}
				}
			}
		}()
	}

	wg.Wait()
}