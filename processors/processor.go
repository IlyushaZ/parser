package processors

import (
	"github.com/IlyushaZ/parser/models"
	"github.com/IlyushaZ/parser/storage"
	"time"
)

type Task struct {
	url, titlePattern, textPattern string
	newsID                         int
}

type Processor struct {
	websiteRepo storage.WebsiteRepository
	newsRepo    storage.NewsRepository
}

func NewProcessor(websiteRepo storage.WebsiteRepository, newsRepo storage.NewsRepository) Processor {
	return Processor{
		websiteRepo: websiteRepo,
		newsRepo:    newsRepo,
	}
}

func (p Processor) ProcessWebsites(tasks chan<- Task) {
	websites := p.websiteRepo.GetUnprocessed()
	if len(websites) == 0 {
		time.Sleep(1 * time.Minute)
	}

	for i := range websites {
		task := Task{
			titlePattern: websites[i].TitlePattern,
			textPattern:  websites[i].TextPattern,
			newsID:       websites[i].ID,
		}

		for _, l := range ScrapLinks(websites[i].MainURL, websites[i].URLPattern) {
			task.url = l
			tasks <- task
		}

		websites[i].Update()
		p.websiteRepo.Update(websites[i])
	}
}

func (p Processor) ProcessNews(tasks <-chan Task) {
	for t := range tasks {
		if p.newsRepo.NewsExists(t.url) {
			continue
		}

		title, text := ScrapNews(t.url, t.titlePattern, t.textPattern)
		news := models.NewNews(t.newsID, t.url, title, text)
		p.newsRepo.Insert(news)
	}
}
