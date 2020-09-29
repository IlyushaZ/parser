package processors

import (
	"github.com/IlyushaZ/parser/models"
	"log"
	"time"
)

type WebsiteRepository interface {
	GetUnprocessed() ([]models.Website, error)
	Update(website *models.Website) error
}

type NewsRepository interface {
	NewsExists(url string) (bool, error)
	Insert(news models.News) error
}

type Task struct {
	url, titlePattern, textPattern string
	websiteID                      int
}

type Processor struct {
	websiteRepo WebsiteRepository
	newsRepo    NewsRepository
}

func NewProcessor(websiteRepo WebsiteRepository, newsRepo NewsRepository) Processor {
	return Processor{
		websiteRepo: websiteRepo,
		newsRepo:    newsRepo,
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
			exists, err := p.newsRepo.NewsExists(l)
			if err != nil {
				log.Println(err)
			}

			if exists {
				continue
			}

			task.url = l
			tasks <- task
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
