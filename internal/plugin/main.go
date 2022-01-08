package plugin

import (
	"errors"
	plg "plugin"
)

const CRAWLER_PLUGIN_NAME = "CrawlerPlugin"

func GetCrawlerPlugin(path string) (*CrawlerPlugin, error) {

	p, err := plg.Open(path)
	if err != nil {
		return nil, err
	}

	var symbol plg.Symbol
	symbol, err = p.Lookup(CRAWLER_PLUGIN_NAME)
	if err != nil {
		return nil, err
	}

	crPlugin, ok := symbol.(*CrawlerPlugin)
	if ok {
		return crPlugin, nil
	}
	return nil, errors.New("GetCrawlerPlugin: could not cast symbol to CrawlerPlugin")

}

func GetCrawlerPlugins(paths ...string) []*CrawlerPlugin {
	res := make([]*CrawlerPlugin, 0, len(paths))

	for _, path := range paths {
		p, err := GetCrawlerPlugin(path)
		if err == nil {
			res = append(res, p)
		}
	}

	return res
}
