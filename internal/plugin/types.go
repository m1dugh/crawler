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

	// the plugin name
	PluginName string

	// PluginEntries for Attachements
	Entries []*CrawlerPluginEntry

	// a map containing additional should add filters
	Filters map[string]*crawler.ShouldAddFilter
}
