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

// creating a non-blocking channel that will trigger the stop
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

	// a channel fed with the added PageRequests (out channel)  
	OnUrlFound     chan []PageRequest
	
	// A non-blocking channel to trigger the stop of the current scan (in channel)
	OnEndRequested chan bool
}
```

### Options

*Options struct:*
```golang
type Options struct {
	// the number of concurrent threads
	MaxWorkers uint

	// the shouldAddFilter for the crawler
	ShouldAddFilter

	// the cookies to add to each request
	Cookies *http.CookieJar

	// the request timeout
	Timeout time.Duration

	// a function providing headers for the request to be made
	HeadersProvider func(PageRequest) http.Header
}
```

*creating basic options:*
```golang
import (
	// ...
	"github.com/m1dugh/crawler"
)

// initializes an option struct with default parameters
var options *crawler.Options = crawler.NewCrawlerOptions()
```


### Scope
A json-compatible structure representing the scope the crawler should be looking into. It is composed of regexes filtering the `url`, the `Content-Type` and the `Extension` of the required file.

*RegexScope struct:*

```golang
type RegexScope struct {
	Includes []string `json:"includes"`
	Excludes []string `json:"excludes"`
}
```

*Scope structure:*

```golang

type Scope struct {
	Urls         *RegexScope `json:"urls"`
	ContentTypes *RegexScope `json:"content-type"`
	Extensions   *RegexScope `json:"extensions"`
}
```



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

### Advanced scopes
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



### CrawlerData
> `CrawlerData` is a struct used to store urls to fetch and fetched urls for the crawler

```golang
type CrawlerData struct {
	// the remaining urls to fetch can both grow and shrink
	UrlsToFetch []PageRequest           `json:"urls_to_fetch"`

	// the PageResults in map whose keys are PageResult.Url.BaseUrl (endpoint of url)
	// and values are all the PageResults which BaseUrl is the same as the key
	FetchedUrls map[string][]PageResult `json:"fetched_urls"`
}
```

*retrieving CrawlerData:*

```golang
var cr *crawler.Crawler;
// ... crawler ends

// returns a copy of crawler data
var data crawler.CrawlerData = cr.GetData()

data["http://any.domain.com/any/endpoint/"] // returns all the PageResults associated to this endpoint
```




### PageRequest

*page request struct :(json-compatible)*
```golang
type PageRequest struct {
	// the endpoint 
	BaseUrl    string            `json:"base_url"`
	// the query parameters
	Parameters map[string]string `json:"params"`
	// the anchor if present
	Anchor     string            `json:"anchor"`
}
```

*useful functions :*
```golang

var url string = "https://www.google.com/#test?q=test"

// converts a url to a PageRequest
var pageRequest PageRequest = PageRequestFromUrl(url)
/*
pageRequest = {
	BaseUrl: "https://www.google.com/",
	Parameters: {"q": "test"},
	Anchor: "test",
}

*/
// converts a page request to a full url
var rebuiltUrl string = pageRequest.ToUrl();
url == rebuiltUrl // returns true

```

### PageResult

> PageResult is a structure created after a page has been fetched

*PageResult struct:*
```golang
type PageResult struct {
	Url           PageRequest   `json:"url"`
	StatusCode    int           `json:"status_code"`
	ContentLength int           `json:"content_length"`
	Headers       http.Header   `json:"headers"`

	// all the urls found by crawling the page
	// get: (*PageResult).FoundUrls
	// set: (*PageResult).SetFoundUrls
	foundUrls []PageRequest
}
```

### ShouldAddFilter

> ShouldAddFilters are functions taking [PageRequest](#pagerequest) as parameters and [CrawlerData](#crawlerdata) returning `true` if the `URL` should be fetched by the crawler

```golang
type ShouldAddFilter func(foundUrl PageRequest, data *CrawlerData) bool
```

*there are three provided ShouldAddFilters:*
- `AggressiveShouldAddFilter` :  only returns false if the url with same parameters and anchor has already been fetched


- `ModerateShouldAddFilter` : returns false if the same endpoint (without parameters and anchors) has responded with the same content-length thrice 

- `LightShouldAddFilter` : returns false if the same endpoint has already been fetched


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

