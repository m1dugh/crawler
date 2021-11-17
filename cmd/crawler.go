package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/akamensky/argparse"
	"github.com/m1dugh/crawler"
)

const DB_FILE_NAME = ".go-crawler.db"

func main() {

	parser := argparse.NewParser("go-crawler", "page crawler for go")

	urls := parser.StringList("u", "url", &argparse.Options{
		Help:     "the list of urls to fetch",
		Required: true,
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

	if err := parser.Parse(os.Args); err != nil {
		log.Fatal("could not parse args: ", err)
	}

	options := &crawler.Options{
		MaxWorkers: uint(*max_workers),
	}

	switch *policy {
	case "AGGRESSIVE", "A":
		options.Policy = crawler.AGGRESSIVE
	case "LIGHT", "L":
		options.Policy = crawler.LIGHT
	default:
		options.Policy = crawler.MODERATE

	}

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
	cr.OnUrlFound = func(req crawler.PageRequest) { fmt.Println(req.ToUrl()) }

	var dbFile *os.File

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

			cr.ResumeScan(data)
		}
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		done <- true
	}()

	cr.OnEndRequested = done
	cr.Crawl(*urls)
	if !cr.Done {
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
