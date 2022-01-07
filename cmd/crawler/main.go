package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/akamensky/argparse"
	"github.com/m1dugh/crawler/internal/config"
	"github.com/m1dugh/crawler/pkg/crawler"
	"github.com/m1dugh/crawler/pkg/plugin"
)

const DB_FILE_NAME = ".go-crawler.db"

func main() {

	parser := argparse.NewParser("go-crawler", "page crawler for go")

	urls := parser.StringList("u", "url", &argparse.Options{
		Help:     "the list of urls to fetch",
		Required: true,
	})

	headers := parser.StringList("H", "header", &argparse.Options{
		Help: "the list of headers to add (\"Header-Key: value1;value2\")",
	})

	requestRate := parser.Int("", "limit", &argparse.Options{
		Help:    "the max requests per seconds",
		Default: -1,
	})

	scopeFile := parser.File("s", "scope", 0, 0, &argparse.Options{
		Help:     "the scope for the crawler",
		Required: true,
	})

	max_workers := parser.Int("t", "threads", &argparse.Options{
		Required: false,
		Default:  10,
		Help:     "the number of max concurrent threads",
	})

	dbFileStr := parser.String("", "resume", &argparse.Options{
		Help:     "the data file of paused scan",
		Required: false,
	})

	policy := parser.Selector("p", "policy", []string{
		"AGGRESSIVE",
		"A",
		"MODERATE", "M",
		"LIGHT", "L",
	}, &argparse.Options{
		Default: "MODERATE",
		Help:    "the level of scanning",
	})

	shouldFetchRobots := parser.Flag("", "robots", &argparse.Options{
		Help:    "fetch robots.txt file for additional urls",
		Default: false,
	})

	// arg parsing
	if err := parser.Parse(os.Args); err != nil {
		log.Fatal("could not parse args: ", err)
	}

	options := &crawler.Options{
		MaxWorkers: int32(*max_workers),
	}

	if headers != nil && len(*headers) > 0 {
		httpHeader := http.Header{}
		for _, h := range *headers {
			splitData := strings.Split(h, ":")
			if len(splitData) >= 2 {
				httpHeader[splitData[0]] = strings.Split(splitData[1], ";")
			}
		}
		options.HeadersProvider = func(_ crawler.PageRequest) http.Header {
			return httpHeader
		}
	}

	switch *policy {
	case "AGGRESSIVE", "A":
		options.ShouldAddFilter = crawler.AggressiveShouldAddFilter
	case "LIGHT", "L":
		options.ShouldAddFilter = crawler.LightShouldAddFilter
	default:
		options.ShouldAddFilter = crawler.ModerateShouldAddFilter

	}

	options.RequestRate = *requestRate

	options.FetchRobots = *shouldFetchRobots

	if _, err := scopeFile.Stat(); err != nil && errors.Is(err, os.ErrNotExist) {
		log.Fatal("file does not exists: ", err)
	}

	body, err := io.ReadAll(scopeFile)

	if err != nil {
		log.Fatal("could not parse scope: ", err)
	}

	var scope crawler.Scope

	if err = json.Unmarshal(body, &scope); err != nil {
		log.Fatal("could not unmarshall json file: ", err)
	}

	cr := crawler.NewCrawler(&scope, options)
	requests := make(chan []crawler.PageRequest, 10)
	cr.OnUrlFound = requests

	go func() {
		for {
			for _, u := range <-requests {
				fmt.Println(u.ToUrl())
			}
		}
	}()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		done <- true
	}()

	cr.OnEndRequested = done

	cr.GetPluginsForDomain = GetOnPageResultAddedHanler(strings.Contains)

	var dbFile *os.File

	// if stopped scan file specified, start scan with given file and urls otherwise crawls with empty data
	if dbFileStr != nil && len(*dbFileStr) > 0 {
		fmt.Println("having resume file")
		if dbFile, err = os.Open(*dbFileStr); err == nil {
			var data crawler.CrawlerData
			body, err = io.ReadAll(dbFile)
			if err != nil {
				log.Fatal("could not read db file: ", err)
			}

			if err = json.Unmarshal(body, &data); err != nil {
				log.Fatal("could not unmarshall json file: ", err)
			}

			var requests []crawler.PageRequest = make([]crawler.PageRequest, len(*urls))
			for i, u := range *urls {
				requests[i] = crawler.PageRequestFromUrl(u)
			}
			data.UrlsToFetch = append(data.UrlsToFetch, requests...)
			cr.ResumeScan(&data)
		} else {
			cr.Crawl(*urls)
		}
	} else {
		cr.Crawl(*urls)
	}

	if !cr.IsDone() {
		var fileName string
		if len(*dbFileStr) > 0 {
			fileName = *dbFileStr
		} else {
			fileName = DB_FILE_NAME
		}
		fmt.Println("stop requested, saving current scan to", fileName)
		var body []byte
		body, err = json.Marshal(cr.GetData())
		if err != nil || os.WriteFile(fileName, body, 0644) != nil {
			log.Fatal("could not save scan")
		}
		fmt.Println("successfully saved current scan at", fileName)
	} else if dbFileStr != nil && len(*dbFileStr) > 0 {
		// removing db file if scan is done
		os.Remove(*dbFileStr)

	}

}

// params:
//  - validateDomainName:
//		a function taking string to be checked in forst argument and string to check against in second argument
func GetOnPageResultAddedHanler(validateDomainName func(string, string) bool) func(string) []crawler.PluginFunction {
	var res func(string) []crawler.PluginFunction

	crawlerPlugins := config.LoadPluginsFromConfig()

	res = func(domainName string) []crawler.PluginFunction {
		res := make([]crawler.PluginFunction, 0)
		for pluginName, p := range crawlerPlugins {
			for _, entry := range p.Entries {
				if validateDomainName(domainName, entry.DomainName) {
					var function crawler.PluginFunction = func(body []byte, pageResult crawler.PageResult, domainResult crawler.DomainResultEntry, output chan<- plugin.Attachements) {
						attachements := (*entry.OnPageResultAdded)(body, pageResult, domainResult)

						for name, value := range attachements {
							newName := pluginName + "." + name
							attachements[newName] = value
							delete(attachements, name)
						}

					}
					res = append(res, function)
				}
			}
		}

		return res
	}

	return res
}
