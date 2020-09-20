package models

//easyjson:json
type News struct {
	ID        int    `db:"id" json:"id"`
	WebsiteID int    `db:"website_id" json:"website_id"`
	URL       string `db:"url" json:"url"`
	Title     string `db:"title" json:"title"`
	Text      string `db:"text" json:"text"`
}

func NewNews(websiteID int, url, title, text string) News {
	return News{
		WebsiteID: websiteID,
		URL:       url,
		Title:     title,
		Text:      text,
	}
}
