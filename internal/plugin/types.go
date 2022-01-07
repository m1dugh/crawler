package plugin

import "github.com/m1dugh/crawler/internal/crawler"

type CrawlerPluginEntry struct {
	DomainName string
	*crawler.OnPageResultAdded
}

type CrawlerPlugin struct {

	// PluginEntries for Attachements
	Entries []*CrawlerPluginEntry

	// a map containing additional should add filters
	Filters map[string]*crawler.ShouldAddFilter
}

type CrawlerPlugins []*CrawlerPlugin
