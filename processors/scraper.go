package processors

import (
	"github.com/gocolly/colly"
	"regexp"
)

func ScrapLinks(url, urlPattern string) (links []string) {
	c := colly.NewCollector()
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		match, _ := regexp.MatchString(urlPattern, e.Attr("href"))
		if match {
			links = append(links, e.Attr("href"))
		}
	})

	c.Visit(url)

	return
}

func ScrapNews(url, titlePattern, textPattern string) (title, text string) {
	c := colly.NewCollector()

	c.OnHTML(titlePattern, func(e *colly.HTMLElement) {
		title = e.Text
	})

	c.OnHTML(textPattern, func(e *colly.HTMLElement) {
		text = e.Text
	})

	c.Visit(url)
	return
}
