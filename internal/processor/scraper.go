package processor

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
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
		entries[u.String()]++
	})

	_ = c.Visit(mainLink)

	for l := range entries {
		if entries[l] == 1 {
			links = append(links, l)
		}
	}

	return
}

func ScrapNews(websiteURL, titlePattern, textPattern string) (title, text string) {
	c := colly.NewCollector()

	c.OnHTML(titlePattern, func(e *colly.HTMLElement) {
		title += strings.TrimSpace(e.Text)
	})

	c.OnHTML(textPattern, func(e *colly.HTMLElement) {
		text += strings.TrimSpace(e.Text)
	})

	_ = c.Visit(websiteURL)
	return
}
