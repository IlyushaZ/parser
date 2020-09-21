package processors

import (
	"github.com/IlyushaZ/parser/models"
	"github.com/IlyushaZ/parser/storage"
	"time"
)

type Task struct {
	url, titlePattern, textPattern string
	websiteID                      int
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

	task := Task{}
	for i := range websites {
		task.titlePattern = websites[i].TitlePattern
		task.textPattern = websites[i].TextPattern
		task.websiteID = websites[i].ID

		for _, l := range ScrapLinks(websites[i].MainURL, websites[i].URLPattern) {
			if p.newsRepo.NewsExists(l) {
				continue
			}

			task.url = l
			tasks <- task
		}

		websites[i].Update()
		p.websiteRepo.Update(websites[i])
	}
}

func (p Processor) ProcessNews(tasks <-chan Task) {
	for t := range tasks {
		title, text := ScrapNews(t.url, t.titlePattern, t.textPattern)
		news := models.NewNews(t.websiteID, t.url, title, text)
		p.newsRepo.Insert(news)
	}
}
