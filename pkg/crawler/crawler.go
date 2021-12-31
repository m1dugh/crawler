package crawler

import (
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/m1dugh/crawler/internal/crawler"
)

// the default shouldAddFilter
var DEFAULT_SHOULD_ADD_FILTER crawler.ShouldAddFilter = AggressiveShouldAddFilter

// Options structure for Crawler object
type Options struct {
	// the number of concurrent threads
	MaxWorkers uint

	// the shouldAddFilter for the crawler
	crawler.ShouldAddFilter

	// flag representing wether the response cookies should be stored or not
	SaveResponseCookies bool

	// the request timeout
	Timeout time.Duration

	// a function providing headers for the request to be made
	// Cookies must be provided in headersProvider
	HeadersProvider func(crawler.PageRequest) http.Header
}

var DEFAULT_HEADERS_PROVIDER = func(crawler.PageRequest) http.Header {
	return http.Header{
		"User-Agent": []string{"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:47.0)", "Gecko/20100101 Firefox/47.3"},
	}
}

// returns a pointer to a default parameter Option struct
func NewCrawlerOptions() *Options {

	return &Options{
		MaxWorkers:          10,
		ShouldAddFilter:     DEFAULT_SHOULD_ADD_FILTER,
		Timeout:             http.DefaultClient.Timeout,
		HeadersProvider:     DEFAULT_HEADERS_PROVIDER,
		SaveResponseCookies: false,
	}
}

type Crawler struct {
	Scope          *crawler.Scope
	Options        *Options
	data           *crawler.CrawlerData
	OnUrlFound     chan []crawler.PageRequest
	OnEndRequested chan bool
	done           bool
}

func NewCrawler(scope *crawler.Scope, opts *Options) *Crawler {

	if opts == nil {
		opts = NewCrawlerOptions()
	}

	return &Crawler{
		Scope:   scope,
		data:    crawler.NewCrawlerData(),
		Options: opts,
		done:    false,
	}
}

func (c *Crawler) IsDone() bool {
	return c.done
}

// launches the crawler with the given data
func (c *Crawler) ResumeScan(data *crawler.CrawlerData) {
	c.data = data
	c.Crawl([]string{})
}

func (c *Crawler) GetData() crawler.CrawlerData {
	return *(c.data)
}

func (c *Crawler) Crawl(baseUrls []string) {

	if c.Scope == nil {
		log.Fatal("scope is not set")
	}

	for _, v := range baseUrls {
		c.data.UrlsToFetch = append(c.data.UrlsToFetch, crawler.PageRequestFromUrl(v))
	}

	var shouldAddFilter crawler.ShouldAddFilter

	if c.Options.ShouldAddFilter != nil {
		shouldAddFilter = c.Options.ShouldAddFilter
	} else {
		shouldAddFilter = DEFAULT_SHOULD_ADD_FILTER
	}

	var cookieJar http.CookieJar = nil
	if c.Options.SaveResponseCookies {
		cookieJar = http.DefaultClient.Jar
	}

	httpClient := &http.Client{
		Jar:     cookieJar,
		Timeout: c.Options.Timeout,
	}

	inChannel := make(chan crawler.PageRequest)
	outChannel := make(chan crawler.PageResult)
	c.done = false

	var workers int32 = 0

	for len(c.data.UrlsToFetch) > 0 || workers > 0 {

		if c.OnEndRequested != nil {
			select {
			case <-c.OnEndRequested:
				return
			default:

			}

		}

		addedWorkers := 0

		for url, ok := c.data.PopUrlToFetch(); c.Options.MaxWorkers-uint(workers) > 0 && ok; url, ok = c.data.PopUrlToFetch() {

			atomic.AddInt32(&workers, 1)
			addedWorkers++
			go func(fetchedUrls map[string]*crawler.DomainResults) {
				defer atomic.AddInt32(&workers, -1)
				url := <-inChannel

				var res crawler.PageResult
				if c.Options.HeadersProvider != nil {
					request, _ := http.NewRequest("GET", url.ToUrl(), nil)
					request.Header = c.Options.HeadersProvider(url)
					res, _ = FetchPage(httpClient, url, c.Scope, fetchedUrls, request)
				} else {
					res, _ = FetchPage(httpClient, url, c.Scope, fetchedUrls, nil)
				}

				outChannel <- res
			}(c.data.FetchedUrls)
			inChannel <- url

		}

		for ; addedWorkers > 0; addedWorkers-- {
			res := <-outChannel
			if len(res.Url.ToUrl()) == 0 {
				continue
			}

			c.data.AddFetchedUrl(res)

			if len(res.FoundUrls) <= 0 {
				continue
			}

			addedUrls := c.data.AddUrlsToFetch(res.FoundUrls, shouldAddFilter, c.Scope)
			if c.OnUrlFound != nil {
				c.OnUrlFound <- addedUrls
			}

		}

	}
	c.done = true
}

func AggressiveShouldAddFilter(foundUrl crawler.PageRequest, data *crawler.CrawlerData) bool {

	fetchedUrls, present := data.FetchedUrls[foundUrl.BaseUrl]

	if !present {
		return true
	}

	for _, urls := range fetchedUrls.Results {
		for _, url := range urls {
			if url.Url.Equals(foundUrl) {
				return false
			}
		}
	}

	return true

}

// has to be over 0
const VALIDITY_COUNT uint8 = 3

func ModerateShouldAddFilter(foundUrl crawler.PageRequest, data *crawler.CrawlerData) bool {

	baseUrl := foundUrl.BaseUrl
	domainName := crawler.ExtractDomainName(baseUrl)
	_, present := data.FetchedUrls[domainName]

	if !present {
		return true
	}

	var fetchedUrls []crawler.PageResult
	fetchedUrls, present = data.FetchedUrls[domainName].Results[baseUrl]

	if !present {
		return true
	}

	if len(fetchedUrls) > int(VALIDITY_COUNT) {
		resultLength := fetchedUrls[0].ContentLength
		for _, url := range fetchedUrls[1:] {
			if url.ContentLength != resultLength {
				return true
			}
		}
	}

	return false

}

func LightShouldAddFilter(foundUrl crawler.PageRequest, data *crawler.CrawlerData) bool {
	_, present := data.FetchedUrls[foundUrl.BaseUrl]

	return !present
}