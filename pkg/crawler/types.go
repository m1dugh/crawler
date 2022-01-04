package crawler

import (
	"github.com/m1dugh/crawler/internal/crawler"
)

type PageRequest = crawler.PageRequest
type PageResult = crawler.PageResult
type DomainResults = crawler.DomainResults
type Scope = crawler.Scope
type CrawlerData = crawler.CrawlerData
type ShouldAddFilter = crawler.ShouldAddFilter

var PageRequestFromUrl = crawler.PageRequestFromUrl

func BasicScope(urls *crawler.RegexScope) *crawler.Scope {
	return &crawler.Scope{
		Urls:         urls,
		ContentTypes: &crawler.RegexScope{},
		Extensions:   &crawler.RegexScope{},
	}
}

type DomainResultEntry = crawler.DomainResultEntry
