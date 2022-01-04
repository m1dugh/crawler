package plugin

import "github.com/m1dugh/crawler/internal/crawler"

type OnPageResultAdded func(
	body string,
	pageResults crawler.PageResult,
	domainResults crawler.DomainResultEntry,
) crawler.Attachement

type CrawlerPluginEntry struct {
	DomainName string
	*OnPageResultAdded
}

type CrawlerPlugin struct {
	PluginName string
	Entries    []*CrawlerPluginEntry
}
