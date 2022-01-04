package plugin

import (
	"io/ioutil"
	plg "plugin"
	"strings"
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

func GetCrawlerPlugins(rootFolder string) []*CrawlerPlugin {

	// gets only the files at the root
	files, err := ioutil.ReadDir(rootFolder)
	if err != nil {
		return nil
	}

	res := make([]*CrawlerPlugin, 0, len(files))

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".so") {
			p := GetCrawlerPlugin(rootFolder + "/" + f.Name())
			if p != nil {
				res = append(res, p)
			}
		}
	}

	return res
}
