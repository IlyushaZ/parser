package main

import (
	"flag"
	"github.com/IlyushaZ/parser/handlers"
	"github.com/IlyushaZ/parser/processors"
	"github.com/IlyushaZ/parser/storage"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
)

const defaultDB = "postgresql://root:root@postgres:5432/parser?sslmode=disable"

func main() {
	var dbURL string
	flag.StringVar(&dbURL, "dbURL", defaultDB, "postgres url")
	flag.Parse()

	db, err := configureDB(dbURL)
	if err != nil {
		log.Fatal("error connecting to database: ", err)
	}
	defer db.Close()

	websiteRepo := storage.NewWebsiteRepository(db)
	newsRepo := storage.NewNewsRepository(db)

	processor := processors.NewProcessor(websiteRepo, newsRepo)

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

	websiteHandler := handlers.NewWebsiteHandler(websiteRepo)
	newsHandler := handlers.NewNewsHandler(newsRepo)

	http.HandleFunc("/websites", websiteHandler.HandlePostWebsite)
	http.HandleFunc("/news", newsHandler.HandleGetNews)
	http.HandleFunc("/news/search", newsHandler.HandleSearchNews)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func configureDB(url string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", url)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(30)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(time.Minute * 2)

	return db, nil
}
