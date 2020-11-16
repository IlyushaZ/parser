package model

type News struct {
	ID        int    `db:"id"`
	WebsiteID int    `db:"website_id"`
	URL       string `db:"url"`
	Title     string `db:"title"`
	Text      string `db:"text"`
}

func NewNews(websiteID int, url, title, text string) News {
	return News{
		WebsiteID: websiteID,
		URL:       url,
		Title:     title,
		Text:      text,
	}
}
