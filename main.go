package crawler

import (
	"log"
	"sync/atomic"
)

// the default shouldAddFilter
var DEFAULT_SHOULD_ADD_FILTER ShouldAddFilter = AggressiveShouldAddFilter

type Options struct {
	// the number of concurrent threads
	MaxWorkers uint

	// the shouldAddFilter for the crawler
	ShouldAddFilter
}

// returns a pointer to a default parameter Option struct
func NewCrawlerOptions() *Options {
	return &Options{
		MaxWorkers:      10,
		ShouldAddFilter: DEFAULT_SHOULD_ADD_FILTER,
	}
}

type Crawler struct {
	Scope          *Scope
	Options        *Options
	data           *CrawlerData
	OnUrlFound     chan []PageRequest
	OnEndRequested chan bool
	done           bool
}

func NewCrawler(scope *Scope, opts *Options) *Crawler {

	if opts == nil {
		opts = NewCrawlerOptions()
	}

	return &Crawler{
		Scope:   scope,
		data:    newCrawlerData(),
		Options: opts,
		done:    false,
	}
}

func (c *Crawler) IsDone() bool {
	return c.done
}

// launches the crawler with the given data
func (c *Crawler) ResumeScan(data *CrawlerData) {
	c.data = data
	c.Crawl([]string{})
}

func (c *Crawler) GetData() CrawlerData {
	return *(c.data)
}

func (c *Crawler) Crawl(baseUrls []string) {

	if c.Scope == nil {
		log.Fatal("scope is not set")
	}

	for _, v := range baseUrls {
		c.data.UrlsToFetch = append(c.data.UrlsToFetch, PageRequestFromUrl(v))
	}

	var shouldAddFilter ShouldAddFilter

	if c.Options.ShouldAddFilter != nil {
		shouldAddFilter = c.Options.ShouldAddFilter
	} else {
		shouldAddFilter = DEFAULT_SHOULD_ADD_FILTER
	}

	inChannel := make(chan PageRequest)
	outChannel := make(chan PageResult)
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

		for url, ok := c.data.popUrlToFetch(); c.Options.MaxWorkers-uint(workers) > 0 && ok; url, ok = c.data.popUrlToFetch() {

			atomic.AddInt32(&workers, 1)
			addedWorkers++
			go func(fetchedUrls map[string][]PageResult) {
				defer atomic.AddInt32(&workers, -1)
				url := <-inChannel
				res, _ := FetchPage(url, *c.Scope, fetchedUrls)

				outChannel <- res
			}(c.data.FetchedUrls)
			inChannel <- url

		}

		for ; addedWorkers > 0; addedWorkers-- {
			res := <-outChannel
			if len(res.Url.ToUrl()) == 0 {
				continue
			}

			c.data.addFetchedUrl(res)

			if len(res.FoundUrls) <= 0 {
				continue
			}

			addedUrls := c.data.addUrlsToFetch(res.FoundUrls, shouldAddFilter, c.Scope)
			if c.OnUrlFound != nil {
				c.OnUrlFound <- addedUrls
			}

		}

	}
	c.done = true
}

func AggressiveShouldAddFilter(foundUrl PageRequest, data *CrawlerData) bool {

	fetchedUrls, present := data.FetchedUrls[foundUrl.BaseUrl]

	if !present {
		return true
	}

	for _, url := range fetchedUrls {
		if url.Url.Equals(foundUrl) {
			return false
		}
	}

	return true

}

// has to be over 0
const VALIDITY_COUNT uint8 = 3

func ModerateShouldAddFilter(foundUrl PageRequest, data *CrawlerData) bool {

	fetchedUrls, present := data.FetchedUrls[foundUrl.BaseUrl]

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

func LightShouldAddFilter(foundUrl PageRequest, data *CrawlerData) bool {
	_, present := data.FetchedUrls[foundUrl.BaseUrl]

	return !present
}
