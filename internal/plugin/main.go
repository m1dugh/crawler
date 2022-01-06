package plugin

import (
	plg "plugin"
)

const CRAWLER_PLUGIN_NAME = "CrawlerPlugin"

func GetCrawlerPlugin(path string) *CrawlerPlugin {

	p, err := plg.Open(path)
	if err != nil {
		return nil
	}

	var symbol plg.Symbol
	symbol, err = p.Lookup(CRAWLER_PLUGIN_NAME)
	if err != nil {
		return nil
	}

	crPlugin, ok := symbol.(*CrawlerPlugin)
	if ok {
		return crPlugin
	}
	return nil
}

func GetCrawlerPlugins(paths ...string) []*CrawlerPlugin {
	res := make([]*CrawlerPlugin, 0, len(paths))

	for _, path := range paths {
		p := GetCrawlerPlugin(path)
		if p != nil {
			res = append(res, p)
		}
	}

	return res
}
