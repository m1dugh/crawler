package crawler

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// a struct representing a request url
//  PageRequest.BaseUrl: the url without any parameter nor anchors
//  PageRequest.Parameters: the parameters
//  PageRequest.Anchor: the anchor
type PageRequest struct {
	BaseUrl    string            `json:"base_url"`
	Parameters map[string]string `json:"params"`
	Anchor     string            `json:"anchor"`
}

func (req *PageRequest) Equals(r2 PageRequest) bool {
	return req.ToUrl() == r2.ToUrl()
}

func (req *PageRequest) getExtensions() string {
	urlParts := strings.Split(req.BaseUrl, "/")
	if len(urlParts[len(urlParts)-1]) <= 0 {
		return "." + strings.Join(strings.Split(urlParts[len(urlParts)-2], ".")[1:], ".")
	}
	return "." + strings.Join(strings.Split(urlParts[len(urlParts)-1], ".")[1:], ".")
}

func (p PageRequest) ToUrl() string {
	var url string = p.BaseUrl
	if len(p.Anchor) > 0 {
		url += "#" + p.Anchor
	}

	if len(p.Parameters) > 0 {
		paramStrings := make([]string, len(p.Parameters))

		i := 0
		for key, param := range p.Parameters {
			paramStrings[i] = fmt.Sprintf("%s=%s", key, param)
			i++
		}

		url += fmt.Sprintf("?%s", strings.Join(paramStrings, "&"))
	}

	return url
}

func PageRequestFromUrl(url string) PageRequest {
	parts := strings.Split(url, "?")
	var req PageRequest
	if len(parts) == 2 {
		req.Parameters = make(map[string]string)
		for _, paramString := range strings.Split(parts[1], "&") {
			data := strings.Split(paramString, "=")
			if len(data) > 1 {
				req.Parameters[data[0]] = data[1]
			} else {
				req.Parameters[data[0]] = ""
			}
		}
	}

	parts = strings.Split(parts[0], "#")
	if len(parts) > 1 {
		req.Anchor = parts[1]
	}

	req.BaseUrl = parts[0]

	return req
}

// a structure representing the response of a page
type PageResult struct {
	Url           PageRequest `json:"url"`
	StatusCode    int         `json:"status_code"`
	ContentLength int         `json:"content_length"`
	Headers       http.Header `json:"headers"`

	// the urls found on the fetched page
	FoundUrls []PageRequest
}

type Attachements struct {
}

func NewAttachements() *Attachements {
	return &Attachements{}
}

type DomainResultEntry struct {
	PageResults   []PageResult `json:"results"`
	*Attachements `json:"attachements"`
}

func NewDomainResultEntry() *DomainResultEntry {
	return &DomainResultEntry{
		PageResults:  make([]PageResult, 0),
		Attachements: NewAttachements(),
	}
}

// a structure grouping PageResults per domain
type DomainResults map[string]*DomainResultEntry

func NewDomainResults() DomainResults {
	return make(DomainResults)
}

type FetchedUrls map[string]DomainResults

type CrawlerData struct {
	UrlsToFetch []PageRequest `json:"urls_to_fetch"`
	FetchedUrls `json:"fetched_urls"`
}

func (results *DomainResults) IsDomainPresent(domainName string) bool {
	_, present := (*results)[domainName]
	return present
}

func NewCrawlerData() *CrawlerData {
	return &CrawlerData{
		make([]PageRequest, 0),
		make(map[string]DomainResults),
	}
}

type ShouldAddFilter func(foundUrl PageRequest, data *CrawlerData) bool

func (d *CrawlerData) AddUrlsToFetch(urls []PageRequest, shouldAdd ShouldAddFilter, scope *Scope) []PageRequest {

	arr := make([]PageRequest, len(urls))
	size := 0
	for _, u := range urls {
		if scope.UrlInScope(u) && d.AddUrlToFetch(u, shouldAdd, scope) {
			arr[size] = u
			size++
		}
	}

	return arr[:size]
}

func (d *CrawlerData) AddUrlToFetch(url PageRequest, shouldAdd ShouldAddFilter, scope *Scope) bool {

	if scope.UrlInScope(url) && shouldAdd(url, d) {
		newArr := FilterArray(append(d.UrlsToFetch, url))
		if len(FilterArray(d.UrlsToFetch)) == len(newArr) {
			return false
		}
		d.UrlsToFetch = newArr
		return true
	}
	return false
}

func (d *CrawlerData) AddFetchedUrl(res PageResult) {

	baseUrl := res.Url.BaseUrl
	domainName := ExtractDomainName(baseUrl)

	if _, present := d.FetchedUrls[domainName]; !present {
		d.FetchedUrls[domainName] = NewDomainResults()
	}

	if results, present := d.FetchedUrls[domainName][baseUrl]; present {
		for _, url := range results.PageResults {
			if url.Url.Equals(res.Url) {
				return
			}
		}

		d.FetchedUrls[domainName][baseUrl].PageResults = append(results.PageResults, res)
	} else {
		if d.FetchedUrls[domainName][baseUrl] == nil {
			d.FetchedUrls[domainName][baseUrl] = NewDomainResultEntry()
		}
		d.FetchedUrls[domainName][baseUrl].PageResults = make([]PageResult, 1)
		d.FetchedUrls[domainName][baseUrl].PageResults[0] = res
	}
}

func (d *CrawlerData) PopUrlToFetch() (PageRequest, bool) {
	if len(d.UrlsToFetch) <= 0 {
		return PageRequest{}, false
	}
	res := d.UrlsToFetch[len(d.UrlsToFetch)-1]

	d.UrlsToFetch = d.UrlsToFetch[:len(d.UrlsToFetch)-1]

	return res, true

}

func (p PageResult) ContentType() string {

	contentType, ok := p.Headers["Content-Type"]
	if !ok || len(contentType) <= 0 {
		return ""
	}

	return strings.Split(contentType[0], ";")[0]
}

type RegexScope struct {
	Includes []string `json:"includes"`
	Excludes []string `json:"excludes"`
}

func (r *RegexScope) matchesRegexScope(value string) bool {

	if len(value) <= 0 {
		return false
	}

	included := false

	if len(r.Includes) == 0 {
		included = true
	}
	for _, include := range r.Includes {
		if r, err := regexp.Compile(include); err == nil && r.MatchString(value) {
			included = true
			break
		}
	}

	if !included {
		return false
	}

	for _, exclude := range r.Excludes {
		if r, err := regexp.Compile(exclude); err == nil && r.MatchString(value) {
			return false
		}
	}

	return true
}

type Scope struct {
	Urls         *RegexScope `json:"urls"`
	ContentTypes *RegexScope `json:"content-type"`
	Extensions   *RegexScope `json:"extensions"`
}

func (s *Scope) UrlInScope(url PageRequest) bool {

	if s.Urls != nil && !s.Urls.matchesRegexScope(url.BaseUrl) {
		return false
	}

	if s.Extensions != nil && !s.Extensions.matchesRegexScope(url.getExtensions()) {
		return false
	}

	return true
}

func (s *Scope) PageInScope(p PageResult) bool {

	if !s.UrlInScope(p.Url) {
		return false
	}

	if s.Extensions != nil && !s.Extensions.matchesRegexScope(p.Url.getExtensions()) {
		return false
	}

	if s.ContentTypes != nil {
		return s.ContentTypes.matchesRegexScope(p.ContentType())
	}

	return true

}

var INCLUDED_MIME_TYPES = []string{
	"text",
	"application/xml",
	"application/x-httpd-php",
	"application/x-sh",
	"application/json",
}
