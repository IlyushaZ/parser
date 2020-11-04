package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/IlyushaZ/parser/handlers"
	"github.com/IlyushaZ/parser/processors"
	"github.com/IlyushaZ/parser/storage"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const defaultDB = "postgresql://root:root@postgres:5432/parser?sslmode=disable"

func main() {
	var dbURL string
	flag.StringVar(&dbURL, "dbURL", defaultDB, "postgres url")
	flag.Parse()

	pool, err := configureDB(dbURL)
	if err != nil {
		log.Fatal("error connecting to database: ", err)
	}
	defer pool.Close()

	websiteRepo := storage.NewWebsiteRepository(pool)
	newsRepo := storage.NewNewsRepository(pool)
	newsCache := storage.NewNewsCache(time.Minute, time.Minute)

	processor := processors.NewProcessor(websiteRepo, newsRepo, newsCache)

	workerChan := make(chan processors.Task)
	for i := 0; i < 5; i++ {
		go processor.ProcessNews(workerChan)
	}

	signal := make(chan struct{})
	go func(p processors.Processor, tasks chan processors.Task, signal chan struct{}) {
		for {
			select {
			case <-signal:
				close(tasks)
				return
			default:
				p.ProcessWebsites(tasks)
			}
		}
	}(processor, workerChan, signal)
	defer close(signal)

	websiteHandler := handlers.NewWebsite(websiteRepo)
	newsHandler := handlers.NewNews(newsRepo)

	http.HandleFunc("/websites", websiteHandler.HandlePostWebsite)
	http.HandleFunc("/news", newsHandler.HandleGetNews)
	http.HandleFunc("/news/search", newsHandler.HandleSearchNews)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func configureDB(url string) (*sqlx.DB, error) {
	pool, err := sqlx.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(); err != nil {
		return nil, err
	}

	pool.SetMaxOpenConns(30)
	pool.SetMaxIdleConns(25)
	pool.SetConnMaxLifetime(time.Minute * 2)

	return pool, nil
}
