package processors

import (
	"log"
	"time"

	"github.com/IlyushaZ/parser/models"
)

type WebsiteRepository interface {
	GetUnprocessed() ([]models.Website, error)
	Update(website *models.Website) error
}

type NewsRepository interface {
	Insert(news models.News) error
}

type NewsCache interface {
	Exists(int, string) bool
	Add(int, string)
}

type Task struct {
	url, titlePattern, textPattern string
	websiteID                      int
}

type Processor struct {
	websiteRepo WebsiteRepository
	newsRepo    NewsRepository
	cache       NewsCache
}

func NewProcessor(websiteRepo WebsiteRepository, newsRepo NewsRepository, cache NewsCache) Processor {
	return Processor{
		websiteRepo: websiteRepo,
		newsRepo:    newsRepo,
		cache:       cache,
	}
}

func (p Processor) ProcessWebsites(tasks chan<- Task) {
	websites, err := p.websiteRepo.GetUnprocessed()
	if err != nil {
		log.Println(err)
	}

	if len(websites) == 0 {
		time.Sleep(1 * time.Minute)
	}

	task := Task{}
	for i := range websites {
		task.titlePattern = websites[i].TitlePattern
		task.textPattern = websites[i].TextPattern
		task.websiteID = websites[i].ID

		for _, l := range ScrapLinks(websites[i].MainURL, websites[i].URLPattern) {
			if p.cache.Exists(task.websiteID, l) {
				continue
			}

			task.url = l
			tasks <- task

			p.cache.Add(task.websiteID, l)
		}

		websites[i].Update()
		if err := p.websiteRepo.Update(&websites[i]); err != nil {
			log.Println(err)
		}
	}
}

func (p Processor) ProcessNews(tasks <-chan Task) {
	for t := range tasks {
		title, text := ScrapNews(t.url, t.titlePattern, t.textPattern)
		news := models.NewNews(t.websiteID, t.url, title, text)

		if err := p.newsRepo.Insert(news); err != nil {
			log.Println(err)
		}
	}
}
