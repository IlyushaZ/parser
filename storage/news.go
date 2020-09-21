package storage

import (
	"github.com/IlyushaZ/parser/models"
	"github.com/jmoiron/sqlx"
	"log"
	"strconv"
)

type NewsRepository interface {
	NewsExists(url string) bool
	Get(limit, offset int) []models.News
	SearchByTitle(search string) []models.News
	Insert(news models.News) error
}

type newsRepository struct {
	db *sqlx.DB
}

func NewNewsRepository(db *sqlx.DB) NewsRepository {
	return newsRepository{db: db}
}

func (nr newsRepository) NewsExists(url string) bool {
	const stmt = "SELECT EXISTS(SELECT 1 FROM news WHERE url = $1)"
	var exists bool
	if err := nr.db.QueryRow(stmt, url).Scan(&exists); err != nil {
		log.Println("storage: error checking if news exists: " + err.Error())
	}

	return exists
}

func (nr newsRepository) Get(limit, offset int) []models.News {
	result := make([]models.News, 0, limit)
	stmt := "SELECT * FROM news LIMIT " + strconv.Itoa(limit) + " OFFSET " + strconv.Itoa(offset)

	rows, err := nr.db.Queryx(stmt)
	if err != nil {
		log.Println("storage: error selecting news: " + err.Error())
	}

	var news models.News
	for rows.Next() {
		rows.StructScan(&news)
		result = append(result, news)
	}

	return result
}

func (nr newsRepository) SearchByTitle(search string) []models.News {
	const stmt = "SELECT * FROM news WHERE title LIKE $1"
	result := make([]models.News, 0)

	rows, err := nr.db.Queryx(stmt, "%"+search+"%")
	if err != nil {
		log.Println("storage: error searching by title: " + err.Error())
		return result
	}

	var news models.News
	for rows.Next() {
		rows.StructScan(&news)
		result = append(result, news)
	}

	return result
}

func (nr newsRepository) Insert(news models.News) error {
	const stmt = "INSERT INTO news (website_id, url, title, text) VALUES ($1, $2, $3, $4)"
	_, err := nr.db.Exec(stmt, news.WebsiteID, news.URL, news.Title, news.Text)

	if err != nil {
		log.Println("error inserting news: " + err.Error())
	}
	return err
}
