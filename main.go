package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	link "go-linkParse"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

/*

	1. GET the webpage
	2. Parse all the links on the page
	3. build the urls with our links
	4. filter out links with a diff domain
	5. find all pages (BFS)
	6. print out XML

*/
const xmlns = "http://www.sitemaps.org/schemas/sitemap/0.9"

type loc struct {
	Value string `xml: "loc"`
}

type urlset struct {
	Urls []loc   `xml:"url"`
	Xmlns string `xml:"xmlns,attr"`
}

func main() {
	urlFlag := flag.String("url", "https://gophercises.com", "the url you want to build a sitemap for")
	maxDepth := flag.Int("depth", 10, "the maximum depth to traverse with bfs")
	flag.Parse()

	pages := bfs(*urlFlag, *maxDepth)
	toXml := urlset{
		Xmlns: xmlns,
	}
	for _, page := range pages {
		toXml.Urls = append(toXml.Urls, loc{page})
	}

	fmt.Println(xml.Header)
	enc := xml.NewEncoder(os.Stdout)
	enc.Indent("", "     ")
	if err := enc.Encode(toXml); err != nil {
		panic(err)
	}
	fmt.Println()
}

func bfs(urlStr string, maxDepth int) []string {
	seen := make(map[string]struct{})
	var q map[string]struct{}
	nq := map[string]struct{} {
		urlStr: struct{}{},
	}
	for i := 0; i <= maxDepth; i++ {
		q, nq = nq, make(map[string]struct{})
		for url, _ := range q {
			// val, ok := seen[ok]
			if _, ok := seen[url]; ok {
				continue
			}
			seen[url] = struct{}{}
			for _, link := range get(url) {
				if _, ok := seen[link]; !ok {
					nq[link] = struct{}{}
				}
			}
		}
	}
	ret := make([]string, 0, len(seen))
	for url, _ := range seen {
		ret = append(ret, url)
	}
	return ret
}

func get(urlString string) []string {
	resp, err := http.Get(urlString)
	if err != nil {
		// panic(err)
		return []string{}
	}
	defer resp.Body.Close()

	/*

		cases

		/some path

		# fragment

	*/

	reqUrl := resp.Request.URL
	baseUrl := &url.URL{
		Scheme:     reqUrl.Scheme,
		Host:       reqUrl.Host,
	}
	base := baseUrl.String()

	return filter(hrefs(resp.Body, base), withPrefix(base))
}

func hrefs(r io.Reader, base string) []string {
	links, _ := link.Parse(r)
	var ret []string
	for _, l := range links {
		switch {
		case strings.HasPrefix(l.Href, "/"):
			ret = append(ret, base + l.Href)
		case strings.HasPrefix(l.Href, "http"):
			ret = append(ret, l.Href)
		}
	}
	return ret
}

func filter(links []string, keepFn func(string) bool) []string{
	var ret []string
	for _, link := range links {
		if keepFn(link) {
		// if strings.HasPrefix(link, base) {
			ret = append(ret, link)
		}
	}
	return ret
}

func withPrefix(pfx string) func(string) bool {
	return func (link string) bool {
		return strings.HasPrefix(link, pfx)
	}
}