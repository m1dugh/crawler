# go-crawler
## a page crawler in go


## Install 

 - #### As Go Module

*`go get` in your shell*
```bash
> go get github.com/m1dugh/crawler
```
*in your code:*
```golang
import "github.com/m1dugh/crawler"

// create a new Crawler
var cr crawler.Crawler;

```
*refer to the [coding documentation](#1-coding-documentation) for further info*

 - #### As command line tool
*download the last version of the executable for your system and refer to the [How To Use Cli](#2-how-to-use-cli) section for further info*

_____
## Documentation

## 1. Coding Documentation


### Getting Started

#### creating a basic Crawler:
```golang

// scope includes any url starting by https://www.google.com
var scope* crawler.Scope = crawler.BasicScope(
	&crawler.RegexScope{
		Includes: [
			"^https://www.google.com"
		],
	}
)

// will generate a crawler with default options and given scope
var cr *crawler.Crawler = crawler.NewCrawler(scope, nil)
var baseUrls = []string{"https://www.google.com/search?q=test"}

// crawler will start crawling 
cr.Crawl(baseUrls)
```

#### stopping a running crawler:
```golang
import (
	// ...
	"time"
	"github.com/m1dugh/crawler"
)

var baseUrls = []string{"https://www.google.com/Search?q=test"}
var cr *crawler.Crawler;
// ... crawler initialization

// the end channel has to be non blocking
var end = make(chan bool)

cr.OnEndRequested = end

// stopping the crawl after 3 seconds 
go func() {
	time.Sleep(3 * time.Second)
	end <- true
}()
cr.Crawl(baseUrls)

// executes after the scan has been stopped
// ...
```


## Types

### Crawler
The crawler object:

*Crawler struct:*
```golang 
type Crawler struct {
	// a pointer to a scope structure
	Scope          *Scope

	// a pointer to a Options structure
	Options        *Options

	// function to call when a url is found 
	OnUrlFound     func(PageRequest)
	
	// A non-blocking channel to trigger the stop of the current scan
	OnEndRequested chan bool
}
```

### Scope
A json-compatible structure representing the scope the crawler should be looking into. It is composed of regexes filtering the `url`, the `Content-Type` and the `Extension` of the required file.

*RegexScope struct:*

```golang
type RegexScope struct {
	Includes: []string
	Excludes: []string
}
```

*Scope structure:*
```golang
var scope *crawler.Scope = &crawler.Scope {
	// crawler will fetch all files whose url matches One of the includes but none of the Excludes
	Urls: &crawler.RegexScope{
		Includes: [
			"^https://www.google.com",
		],
		Excludes: [
			"^https://www.google.com/search",
		],
	},
	// crawler will parse all files whose extension matches One of the includes but none of the Excludes
	Extensions: &crawler.RegexScope{
		Includes: [
			"\.\w+ml",
		],
		Excludes: [
			"\.xml",
		],
	},
	// crawler will parse all files whose ContentType matches One of the includes but none of the Excludes
	ContentType: &crawler.RegexScope{
		Includes: [
			"^text",
		],
		Exludes: [
			"^text/plain$"
		],
	}
}
```

*Example :*
|                    URL                   | fetched | parsed |reason|
|:-----------------------------------------|:--------|:-------|:-----|
| https://www.google.com/images/data.html (content-type starts with `text` and not `text/plain`)  |  true   |  true  | -    |
| https://www.google.com/search/data.html  | false   | false   | url  |
| https://www.google.com/images/data.xml   | false   | false  |extension|
| https://www.google.com/images/data.php   | false   | false | extension |
| https://www.google.com/images/data.xhtml (content-type: text/plain) | true | false | content-type |
| https://www.google.com/images/data.html (content-type: application/json) | false | false | content-type |

*NB: An Empty `crawler.RegexScope` will result in assuming all assumptions are correct; the following will only filter based on the `url` regexes given*

### creating a basic scope
```golang
var scope* crawler.Scope = &crawler.Scope{
	Urls: &crawler.RegexScope{
		Includes: [
			"any_url",
		],
		Excludes: [
			"any_exclusion",
		],
	},
	Extensions: &crawler.RegexScope{},
	ContentType: &crawler.RegexScope{},
}

// shorthand for the above:
var scope* crawler.Scope = crawler.BasicScope(&crawler.RegexScope{
	Includes: [
			"any_url",
	],
	Excludes: [
		"any_exclusion",
	],
})
```

### ShouldAddFilter

> ShouldAddFilters are functions taking [PageRequest](#pagerequest) as parameters and [CrawlerData](#crawlerdata) returning `true` if the `URL` should be fetched by the crawler

```golang
type ShouldAddFilter func(foundUrl PageRequest, data *CrawlerData) bool
```

*three provided ShouldAddFilters:*
 - `AggressiveShouldAddFilter`
```golang
// only returns false if the url with same parameters and anchor has already been fetched
func AggressiveShouldAddFilter(foundUrl PageRequest, data *CrawlerData) bool {

	fetchedUrls, present := data.FetchedUrls[foundUrl.BaseUrl]

	if !present {
		return true
	}

	for _, url := range fetchedUrls {
		
		if url.Url.Equals(foundUrl) {
			return false
		}
	}

	return true

}
```

- `ModerateShouldAddFilter`

```golang
const VALIDITY_COUNT uint8 = 3
// returns false if the same endpoint (without parameters and anchors) has responded with the same content-length thrice 
func ModerateShouldAddFilter(foundUrl PageRequest, data *CrawlerData) bool {

	fetchedUrls, present := data.FetchedUrls[foundUrl.BaseUrl]

	if !present {
		return true
	}

	if len(fetchedUrls) > int(VALIDITY_COUNT) {
		resultLength := fetchedUrls[0].ContentLength
		for _, url := range fetchedUrls[1:] {
			if url.ContentLength != resultLength {
				return true
			}
		}
	}

	return false

}
```

 - `LightShouldAddFilter`
```golang
// returns false if the same endpoint has already been fetched
func LightShouldAddFilter(foundUrl PageRequest, data *CrawlerData) bool {
	_, present := data.FetchedUrls[foundUrl.BaseUrl]

	return !present
}
```


## 2. How to use CLI

### parameters

> `--url|-u url[,url]` : required: the url(s) to first crawl at

> `--scope|-s scope`: required: the path to the scope file (see below)

> `--resume dbFile`: the path to a db file of an older scan. if not found, the scan will start from scratch. If the scan is stopped, the current scan will be stored in the file specified

> `--threads|-t int`: the numbers of concurrent threads crawling together (default is 10)

> `--policy|-p {LIGHT, MODERATE, AGGRESSIVE}`: the crawling policy (default: `MODERATE`). for further information, see [should add filters](#shouldaddfilter)

### basic crawling

*scope.json*
```json
{
	"urls": {
		"includes": [
			"https://(\w+\.)*www.google.com"
		],
		"excludes": [
			"\/search"
		]
	}
}
```

the `scope.json` file is a file representing the scope the crawler is authorized to crawl in, above is a small example of a scope crawling all pages belonging to any subdomain of `www.google.com` and with url not containing `/search`.
> The scope file is a file that acts like the [`crawler.Scope`](#scope) structure for the cli. Thus, it has `urls`, `content-type` and `extensions` fields, each containing a `includes` and `excludes` members containing regexes.

```bash
> ./crawler --url https://www.google.com/ --scope ./scope.json
```

> if scan is paused in the console, it will save a `.go-crawler.db` file or whatever name specified in `--resume` flag.

#### resuming a paused scan

```bash
> ./crawler --url any_url --scope scope.json --resume .go-crawler.db
```

