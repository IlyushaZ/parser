package storage

import (
	"strconv"

	"github.com/IlyushaZ/parser/internal/model"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type NewsRepository struct {
	db *sqlx.DB
}

func NewNewsRepository(db *sqlx.DB) NewsRepository {
	return NewsRepository{db: db}
}

func (nr NewsRepository) Get(limit, offset int) (result []model.News, err error) {
	result = make([]model.News, 0, limit)
	stmt := "SELECT * FROM news LIMIT " + strconv.Itoa(limit) + " OFFSET " + strconv.Itoa(offset)

	rows, err := nr.db.Queryx(stmt)
	if err != nil {
		err = errors.WithMessage(err, "news storage: err selecting list of news")
		return
	}

	var news model.News
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

	if len(result) == 0 {
		err = ErrNotFound
	}

	return
}

func (nr NewsRepository) SearchByTitle(search string) (result []model.News, err error) {
	const stmt = "SELECT * FROM news WHERE title LIKE $1"
	result = make([]model.News, 0)

	rows, err := nr.db.Queryx(stmt, "%"+search+"%")
	if err != nil {
		err = errors.WithMessage(err, "news storage: err searching news by title "+search)
		return
	}
	defer rows.Close()

	var news model.News
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

	if len(result) == 0 {
		err = ErrNotFound
	}

	return
}

func (nr NewsRepository) Insert(news model.News) error {
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

func (nr NewsRepository) Exists(url string) (exists bool) {
	const stmt = "SELECT EXISTS(SELECT id FROM NEWS WHERE url = $1)"
	_ = nr.db.QueryRow(stmt, url).Scan(&exists)
	return
}

const defaultTTL = 1800

type NewsCache struct {
	mc  *memcache.Client
	ttl int32
}

func NewNewsCache(mc *memcache.Client, ttl int32) NewsCache {
	if ttl == 0 {
		ttl = defaultTTL
	}

	return NewsCache{
		mc:  mc,
		ttl: ttl,
	}
}

func (nc NewsCache) Exists(url string, websiteID int32) (bool, error) {
	item, err := nc.mc.Get(url)
	if err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) {
			return false, ErrNotFound
		}

		err = errors.WithMessage(err, "news storage: error checking if url exists")
		return false, err
	}

	if string(item.Value) == strconv.Itoa(int(websiteID)) {
		return true, nil
	}

	return false, nil
}

func (nc NewsCache) Add(url string, websiteID int32) error {
	err := nc.mc.Set(&memcache.Item{
		Key:        url,
		Value:      []byte(strconv.Itoa(int(websiteID))),
		Expiration: nc.ttl,
	})

	if err != nil {
		err = errors.WithMessage(err, "news storage: err adding url")
	}

	return err
}
