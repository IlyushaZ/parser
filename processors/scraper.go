package processors

import (
	"github.com/gocolly/colly"
	"net/url"
	"regexp"
	"strings"
)

func ScrapLinks(mainLink, linkPattern string) (links []string) {
	c := colly.NewCollector()
	entries := make(map[string]int)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		raw := e.Attr("href")
		match, _ := regexp.MatchString(linkPattern, raw)

		if !match {
			return
		}

		u, err := url.Parse(raw)
		if err != nil {
			return
		}

		u.RawQuery = ""
		entries[u.String()] += 1
	})

	c.Visit(mainLink)

	for l := range entries {
		if entries[l] == 1 {
			links = append(links, l)
		}
	}

	return
}

func ScrapNews(url, titlePattern, textPattern string) (title, text string) {
	c := colly.NewCollector()

	c.OnHTML(titlePattern, func(e *colly.HTMLElement) {
		title += strings.TrimSpace(e.Text)
	})

	c.OnHTML(textPattern, func(e *colly.HTMLElement) {
		text += strings.TrimSpace(e.Text)
	})

	c.Visit(url)
	return
}
