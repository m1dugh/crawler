package plugin

import (
	plg "plugin"
)

const CRAWLER_PLUGIN_NAME = "CrawlerPlugin"

func GetCrawlerPlugin(paths ...string) []*CrawlerPlugin {
	result := make([]*CrawlerPlugin, 0, len(paths))
	for _, path := range paths {
		p, err := plg.Open(path)
		if err != nil {
			continue
		}

		var symbol plg.Symbol
		symbol, err = p.Lookup(CRAWLER_PLUGIN_NAME)
		if err != nil {
			continue
		}

		crPlugin, ok := symbol.(*CrawlerPlugin)
		if ok {
			result = append(result, crPlugin)
		}
	}

	return result
}
