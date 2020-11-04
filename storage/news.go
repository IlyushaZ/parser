package storage

import (
	"strconv"
	"sync"
	"time"

	"github.com/IlyushaZ/parser/models"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type NewsRepository struct {
	db *sqlx.DB
}

func NewNewsRepository(db *sqlx.DB) NewsRepository {
	return NewsRepository{db: db}
}

func (nr NewsRepository) Get(limit, offset int) (result []models.News, err error) {
	result = make([]models.News, 0, limit)
	stmt := "SELECT * FROM news LIMIT " + strconv.Itoa(limit) + " OFFSET " + strconv.Itoa(offset)

	rows, err := nr.db.Queryx(stmt)
	if err != nil {
		err = errors.WithMessage(err, "news storage: err selecting list of news")
		return
	}

	var news models.News
	for rows.Next() {
		_ = rows.StructScan(&news)
		result = append(result, news)
	}

	if rows.Err() != nil {
		err = errors.WithMessage(
			err,
			"news storage: err scanning news from db to struct while getting list of news",
		)
	}

	return
}

func (nr NewsRepository) SearchByTitle(search string) (result []models.News, err error) {
	const stmt = "SELECT * FROM news WHERE title LIKE $1"
	result = make([]models.News, 0)

	rows, err := nr.db.Queryx(stmt, "%"+search+"%")
	if err != nil {
		err = errors.WithMessage(err, "news storage: err searching news by title "+search)
		return
	}
	defer rows.Close()

	var news models.News
	for rows.Next() {
		_ = rows.StructScan(&news)
		result = append(result, news)
	}

	if rows.Err() != nil {
		err = errors.WithMessage(
			err,
			"news storage: err scanning news from db to struct while searching "+search,
		)
	}

	return
}

func (nr NewsRepository) Insert(news models.News) error {
	const stmt = "INSERT INTO news (website_id, url, title, text) VALUES ($1, $2, $3, $4)"
	_, err := nr.db.Exec(stmt, news.WebsiteID, news.URL, news.Title, news.Text)

	if err != nil {
		err = errors.WithMessagef(
			err, "news storage: err inserting news for website %d with url %s",
			news.WebsiteID,
			news.URL,
		)
	}

	return err
}

type newsCacheItem struct {
	until time.Time
}

type websiteCacheItem struct {
	mu   sync.RWMutex
	news map[string]*newsCacheItem
}

type NewsCache struct {
	duration  time.Duration
	checkFreq time.Duration

	items map[int]*websiteCacheItem

	mu   sync.RWMutex
	once sync.Once
}

func NewNewsCache(duration, checkFreq time.Duration) *NewsCache {
	return &NewsCache{
		duration:  duration,
		checkFreq: checkFreq,
	}
}

func (c *NewsCache) Exists(websiteID int, url string) bool {
	c.internalInit()
	c.mu.RLock()
	defer c.mu.RUnlock()

	if _, ok := c.items[websiteID]; !ok {
		return false
	}

	if _, ok := c.items[websiteID].news[url]; ok {
		return true
	}

	return false
}

func (c *NewsCache) Add(websiteID int, url string) {
	c.internalInit()
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.items[websiteID]; !ok {
		c.items[websiteID] = &websiteCacheItem{news: make(map[string]*newsCacheItem)}
	}

	item := newsCacheItem{}
	if c.duration != 0 {
		item.until = time.Now().Add(c.duration)
	}

	c.items[websiteID].news[url] = &item
}

func (c *NewsCache) internalInit() {
	c.once.Do(func() {
		c.items = make(map[int]*websiteCacheItem)
		c.clear()
	})
}

// TODO: make a limit of goroutines running at the same time
func (c *NewsCache) clear() {
	go func() {
		removeOverdue := func(w *websiteCacheItem, wg *sync.WaitGroup) {
			w.mu.Lock()
			defer wg.Done()
			defer w.mu.Unlock()

			now := time.Now()
			for url, item := range w.news {
				if item.until.Before(now) {
					delete(w.news, url)
				}
			}
		}

		var wg sync.WaitGroup

		for {
			if c.duration == 0 {
				break
			}

			for _, website := range c.items {
				wg.Add(1)
				go removeOverdue(website, &wg)
			}

			wg.Wait()
			time.Sleep(c.checkFreq)
		}
	}()
}
