package crawler

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/m1dugh/crawler/internal/crawler"
)

func FetchPage(httpClient *http.Client, url crawler.PageRequest, scope *crawler.Scope, fetchedUrls map[string]*crawler.DomainResults, request *http.Request) (crawler.PageResult, error) {

	if request == nil {
		request, _ = http.NewRequest("GET", url.ToUrl(), nil)
	}
	res, err := httpClient.Do(request)
	if err != nil {
		return crawler.PageResult{}, err
	}

	result := crawler.PageResult{
		Url:           url,
		StatusCode:    res.StatusCode,
		ContentLength: int(res.ContentLength),
		Headers:       res.Header.Clone(),
		FoundUrls:     make([]crawler.PageRequest, 0)}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return crawler.PageResult{}, err
	}

	shouldExtractUrls := false
	for _, mimeType := range crawler.INCLUDED_MIME_TYPES {
		if strings.HasPrefix(result.ContentType(), mimeType) {
			shouldExtractUrls = true
			break
		}
	}

	if shouldExtractUrls {
		urls := crawler.ExtractUrlsFromHtml(string(body), url.BaseUrl)

		data := make([]crawler.PageRequest, len(urls))

		size := 0
		for _, v := range urls {
			if scope.UrlInScope(v) {
				data[size] = v
				size++
			}
		}

		result.FoundUrls = data[:size]
	}

	return result, nil

}

func FetchRobots(rootUrl string) []crawler.PageRequest {
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
	result := make([]crawler.PageRequest, 0, len(lines))
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

		result = append(result, crawler.PageRequestFromUrl(rootUrl+line))

	}

	return result
}
