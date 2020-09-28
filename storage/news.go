package storage

import (
	"github.com/IlyushaZ/parser/models"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"strconv"
)

type NewsRepository struct {
	db *sqlx.DB
}

func NewNewsRepository(db *sqlx.DB) NewsRepository {
	return NewsRepository{db: db}
}

func (nr NewsRepository) NewsExists(url string) (exists bool, err error) {
	const stmt = "SELECT EXISTS(SELECT 1 FROM news WHERE url = $1)"

	err = nr.db.QueryRow(stmt, url).Scan(&exists)
	if err != nil {
		err = errors.WithMessage(err, "news storage: err checking if news exists with url "+url)
	}

	return
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
