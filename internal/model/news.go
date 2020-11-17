package model

type News struct {
	ID        int32  `db:"id"`
	WebsiteID int32  `db:"website_id"`
	URL       string `db:"url"`
	Title     string `db:"title"`
	Text      string `db:"text"`
}

func NewNews(websiteID int32, url, title, text string) News {
	return News{
		WebsiteID: websiteID,
		URL:       url,
		Title:     title,
		Text:      text,
	}
}
