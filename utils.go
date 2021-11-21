package crawler

import (
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

const rootUrlString = `https?://((\w+\.)+[a-z]{2,5}|localhost|((\d{1,3}\.){3})\d{1,3})(:\d+)?`
const locationString = `(/[^"'\s><\\\*]*)+`

var rootUrlPattern = regexp.MustCompile(rootUrlString)
var urlPattern = regexp.MustCompile(rootUrlString + locationString)
var locationPattern = regexp.MustCompile(fmt.Sprintf(`"%s"`, locationString))

func getProtocol(url string) string {
	return strings.Split(url, "://")[0]
}

/* a function that extracts urls from any page
 */
func extractUrlsFromHtml(page string, url string) []PageRequest {
	foundLinks := make([]PageRequest, 0)

	rootUrl := rootUrlPattern.FindString(url)

	for _, v := range urlPattern.FindAllString(page, -1) {
		foundLinks = append(foundLinks, PageRequestFromUrl(html.UnescapeString(v)))
	}

	for _, loc := range locationPattern.FindAllString(page, -1) {

		if len(loc) <= 2 {
			continue
		}
		loc = loc[1 : len(loc)-1]
		if len(loc) >= 1 {
			var effectiveLink string
			if len(loc) >= 2 && loc[:2] == "//" {
				effectiveLink = getProtocol(url) + ":" + loc
			} else {
				effectiveLink = rootUrl + loc
			}

			effectiveLink = html.UnescapeString(effectiveLink)
			foundLinks = append(foundLinks, PageRequestFromUrl(effectiveLink))
		}
	}

	return filterArray(foundLinks)
}

func filterArray(pages []PageRequest) []PageRequest {
	index := make(map[string]bool)

	size := 0
	for _, req := range pages {
		if _, ok := index[req.ToUrl()]; !ok {
			pages[size] = req
			index[req.ToUrl()] = false
			size++
		}
	}

	return pages[:size]
}

func FetchPage(httpClient *http.Client, url PageRequest, scope *Scope, fetchedUrls map[string][]PageResult, request *http.Request) (PageResult, error) {

	if request == nil {
		request, _ = http.NewRequest("GET", url.ToUrl(), nil)
	}
	res, err := httpClient.Do(request)
	if err != nil {
		return PageResult{}, err
	}

	result := PageResult{url, res.StatusCode, int(res.ContentLength), res.Header.Clone(), make([]PageRequest, 0)}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return PageResult{}, err
	}

	shouldExtractUrls := false
	for _, mimeType := range INCLUDED_MIME_TYPES {
		if strings.HasPrefix(result.ContentType(), mimeType) {
			shouldExtractUrls = true
			break
		}
	}

	if shouldExtractUrls {
		urls := extractUrlsFromHtml(string(body), url.BaseUrl)

		data := make([]PageRequest, len(urls))

		size := 0
		for _, v := range urls {
			if scope.UrlInScope(v) {
				data[size] = v
				size++
			}
		}

		result.SetFoundUrls(data[:size])
	}

	return result, nil

}

func FetchRobots(rootUrl string) []PageRequest {
	res, err := http.Get(rootUrl + "/robots.txt")
	if err != nil {
		return nil
	}

	defer res.Body.Close()
	var body []byte
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(body[:]), "\n")
	result := make([]PageRequest, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(line, "Disallow:") {
			line = line[len("Disallow:"):]
		} else if strings.HasPrefix(line, "Allow:") {
			line = line[len("Allow:"):]
		} else {
			continue
		}

		// remove wildcarded urls
		if strings.Contains(line, "*") {
			continue
		}

		line = strings.ReplaceAll(line, " ", "")

		result = append(result, PageRequestFromUrl(rootUrl+line))

	}

	return result
}
