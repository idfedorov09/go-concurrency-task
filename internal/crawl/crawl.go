package crawl

type Fetcher interface {
	Fetch(url string) (body string, urls []string, err error)
}

type PageFetcher interface {
    FetchPage(url string) (PageInfo, error)
}

type FetchFunc func(targetUrl string) (PageInfo, error)

type crawlSession[T PageFetcher] struct {
    deps T
    fn   FetchFunc
}

type Wrapper[T PageFetcher] func(deps T, next FetchFunc) FetchFunc

type PageInfo struct {
	URL string
	Body string
	URLs []string
}

func CrawlPage[T PageFetcher](targetUrl string, session crawlSession[T]) <-chan string {
	urlsParsed := make(chan string)
	go func() {
		defer close(urlsParsed)

		page, err := session.fetch(targetUrl)
		if err != nil {
			return
		}
		for _, outUrl := range page.URLs {
			urlsParsed <- outUrl
		}
	}()
	return urlsParsed
}

func NewCrawlSession[T PageFetcher](deps T) crawlSession[T] {
    return crawlSession[T]{
        deps: deps,
        fn:   deps.FetchPage,
    }
}

func FetchPage(fetcher Fetcher, targetUrl string) (PageInfo, error) {
	body, urls, err := fetcher.Fetch(targetUrl)
	if err != nil {
		return PageInfo{}, err
	}

	return PageInfo{
		URL:  targetUrl,
		Body: body,
		URLs: urls,
	}, nil
}

func (cs crawlSession[T]) Wrap(wrapper Wrapper[T]) crawlSession[T] {
	cs.fn = wrapper(cs.deps, cs.fn)
	return cs
}

func (s crawlSession[T]) fetch(targetURL string) (PageInfo, error) {
	return s.fn(targetURL)
}
