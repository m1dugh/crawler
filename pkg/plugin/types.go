package plugin

import (
	crawler "github.com/m1dugh/crawler/internal/crawler"
	cr_plugin "github.com/m1dugh/crawler/internal/plugin"
)

type Attachements = crawler.Attachements
type OnPageResultAdded = crawler.OnPageResultAdded

type CrawlerPluginEntry = cr_plugin.CrawlerPluginEntry

type CrawlerPlugin = cr_plugin.CrawlerPlugin

var NewAttachements = crawler.NewAttachements
