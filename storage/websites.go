package storage

import (
	"fmt"
	"github.com/IlyushaZ/parser/models"
	"github.com/jmoiron/sqlx"
	"strconv"
)

const unprocessedLimit = 50

type WebsiteRepository interface {
	GetUnprocessed() []models.Website
	Insert(website models.Website) error
	Update(website models.Website) error
}

type websiteRepository struct {
	db *sqlx.DB
}

func NewWebsiteRepository(db *sqlx.DB) WebsiteRepository {
	return websiteRepository{db: db}
}

func (wr websiteRepository) GetUnprocessed() []models.Website {
	websites := make([]models.Website, 0, unprocessedLimit)
	stmt := "SELECT * FROM websites WHERE process_at < NOW() LIMIT " + strconv.Itoa(unprocessedLimit)

	rows, _ := wr.db.Queryx(stmt)

	var website models.Website
	for rows.Next() {
		_ = rows.StructScan(&website)
		websites = append(websites, website)
	}

	return websites
}

func (wr websiteRepository) Insert(website models.Website) error {
	const stmt = "INSERT INTO websites (main_url, url_pattern, title_pattern, text_pattern) VALUES ($1, $2, $3, $4)"

	_, err := wr.db.Exec(stmt, website.MainURL, website.URLPattern, website.TitlePattern, website.TextPattern)
	return err
}

func (wr websiteRepository) Update(website models.Website) error {
	const stmt = "UPDATE websites SET process_at = $1 WHERE id = $2"
	_, err := wr.db.Exec(stmt, website.ProcessAt, website.ID)

	if err != nil {
		fmt.Println(err)
	}
	return err
}
