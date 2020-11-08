package processor

import (
	"errors"
	"log"
	"time"

	"github.com/IlyushaZ/parser/internal/model"
	"github.com/IlyushaZ/parser/internal/storage"
)

type WebsiteRepository interface {
	GetUnprocessed() ([]model.Website, error)
	Update(*model.Website) error
}

type NewsRepository interface {
	Insert(model.News) error
	Exists(string) bool
}

type NewsCache interface {
	Exists(string, int) (bool, error)
	Add(string, int) error
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

func New(websiteRepo WebsiteRepository, newsRepo NewsRepository, cache NewsCache) Processor {
	return Processor{
		websiteRepo: websiteRepo,
		newsRepo:    newsRepo,
		cache:       cache,
	}
}

func (p Processor) ProcessWebsites(tasks chan<- Task) {
	websites, err := p.websiteRepo.GetUnprocessed()
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		log.Println(err)
	}

	if errors.Is(err, storage.ErrNotFound) {
		time.Sleep(1 * time.Minute)
	}

	task := Task{}
	for i := range websites {
		task.titlePattern = websites[i].TitlePattern
		task.textPattern = websites[i].TextPattern
		task.websiteID = websites[i].ID

		for _, l := range ScrapLinks(websites[i].MainURL, websites[i].URLPattern) {
			exists, err := p.cache.Exists(l, task.websiteID)
			if err != nil && !errors.Is(err, storage.ErrNotFound) {
				log.Println(err)
			}

			// if news exists in cache
			// or if it has gone from cache, but exists in db
			if exists || p.newsRepo.Exists(l) {
				continue
			}

			// new urls are going to cache
			if err = p.cache.Add(l, task.websiteID); err != nil {
				log.Println(err)
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
		news := model.NewNews(t.websiteID, t.url, title, text)

		if err := p.newsRepo.Insert(news); err != nil {
			log.Println(err)
		}
	}
}
