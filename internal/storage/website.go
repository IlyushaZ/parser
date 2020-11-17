package storage

import (
	"strconv"

	"github.com/IlyushaZ/parser/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

const unprocessedLimit = 50

type WebsiteRepository struct {
	db *sqlx.DB
}

func NewWebsiteRepository(db *sqlx.DB) WebsiteRepository {
	return WebsiteRepository{db: db}
}

func (wr WebsiteRepository) GetUnprocessed() (websites []model.Website, err error) {
	websites = make([]model.Website, 0, unprocessedLimit)
	stmt := "SELECT * FROM websites WHERE process_at < NOW() LIMIT " + strconv.Itoa(unprocessedLimit)

	rows, err := wr.db.Queryx(stmt)
	if err != nil {
		err = errors.WithMessage(err, "website storage: err selecting unprocessed websites")
		return
	}

	var website model.Website
	for rows.Next() {
		_ = rows.StructScan(&website)
		websites = append(websites, website)
	}

	if rows.Err() != nil {
		err = errors.WithMessage(err, "website storage: err scanning unprocessed websites")
	}

	if len(websites) == 0 {
		err = ErrNotFound
	}

	return
}

func (wr WebsiteRepository) Insert(website *model.Website) error {
	const stmt = "INSERT INTO websites (main_url, url_pattern, title_pattern, text_pattern) VALUES ($1, $2, $3, $4)"

	_, err := wr.db.Exec(stmt, website.MainURL, website.URLPattern, website.TitlePattern, website.TextPattern)
	if err != nil {
		err = errors.WithMessage(err, "website storage: err inserting website")
	}

	return err
}

func (wr WebsiteRepository) WebsiteExists(url string) (exists bool) {
	const stmt = "SELECT EXISTS(SELECT id FROM websites WHERE main_url = $1)"
	_ = wr.db.QueryRow(stmt, url).Scan(&exists)
	return
}

func (wr WebsiteRepository) Update(website *model.Website) error {
	const stmt = "UPDATE websites SET process_at = $1 WHERE id = $2"

	_, err := wr.db.Exec(stmt, website.ProcessAt, website.ID)
	if err != nil {
		err = errors.WithMessage(
			err,
			"website storage: err updating website with id "+strconv.Itoa(int(website.ID)),
		)
	}

	return err
}
