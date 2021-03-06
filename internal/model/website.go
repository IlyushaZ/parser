package model

import "time"

const processAtFrequency = 5 * time.Minute

type Website struct {
	ID           int32     `db:"id"`
	MainURL      string    `db:"main_url"`
	URLPattern   string    `db:"url_pattern"`
	TitlePattern string    `db:"title_pattern"`
	TextPattern  string    `db:"text_pattern"`
	ProcessAt    time.Time `db:"process_at"`
}

func NewWebsite(mainURL, urlPattern, titlePattern, textPattern string) Website {
	return Website{
		MainURL:      mainURL,
		URLPattern:   urlPattern,
		TitlePattern: titlePattern,
		TextPattern:  textPattern,
		ProcessAt:    time.Now(),
	}
}

func (w *Website) Update() {
	w.ProcessAt = w.ProcessAt.Add(processAtFrequency)
}
