package crawler

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

func FilterArray(pages []PageRequest) []PageRequest {
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

const rootUrlString = `https?://((\w+\.)+[a-z]{2,5}|localhost|((\d{1,3}\.){3})\d{1,3})(:\d+)?`
const locationString = `(/[^"'\s><\\\*]*)+`

var rootUrlPattern = regexp.MustCompile(rootUrlString)
var urlPattern = regexp.MustCompile(rootUrlString + locationString)
var locationPattern = regexp.MustCompile(fmt.Sprintf(`"%s"`, locationString))

func GetProtocol(url string) string {
	return strings.Split(url, "://")[0]
}

/* a function that extracts urls from any page
 */
func ExtractUrlsFromHtml(page string, url string) []PageRequest {
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
				effectiveLink = GetProtocol(url) + ":" + loc
			} else {
				effectiveLink = rootUrl + loc
			}

			effectiveLink = html.UnescapeString(effectiveLink)
			foundLinks = append(foundLinks, PageRequestFromUrl(effectiveLink))
		}
	}

	return FilterArray(foundLinks)
}

func ExtractDomainName(url string) string {
	rootUrl := rootUrlPattern.FindString(url)
	if len(rootUrl) <= 0 {
		return ""
	}
	// removes <protocol>://
	return rootUrl[len(GetProtocol(rootUrl))+3:]
}
