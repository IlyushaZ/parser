package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/IlyushaZ/parser/internal/handler"
	"github.com/IlyushaZ/parser/internal/processor"
	"github.com/IlyushaZ/parser/internal/storage"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	defaultDB       = "postgresql://root:root@postgres:5432/parser?sslmode=disable"
	defaultMemcache = "memcached:11211"
)

func main() {
	var dbURL, mcURL string
	flag.StringVar(&dbURL, "dbURL", defaultDB, "postgres url")
	flag.StringVar(&mcURL, "mcURL", defaultMemcache, "memcache url")
	flag.Parse()

	pool, err := configureDB(dbURL)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	websiteRepo := storage.NewWebsiteRepository(pool)
	newsRepo := storage.NewNewsRepository(pool)

	mc := memcache.New(mcURL)
	if err = mc.Ping(); err != nil {
		panic(err)
	}
	mc.MaxIdleConns = 10
	if err = mc.DeleteAll(); err != nil {
		panic(err)
	}
	newsCache := storage.NewNewsCache(mc, 1800)

	proc := processor.New(websiteRepo, newsRepo, newsCache)

	workerChan := make(chan processor.Task)
	for i := 0; i < 5; i++ {
		go proc.ProcessNews(workerChan)
	}

	signal := make(chan struct{})
	go func(p processor.Processor, tasks chan processor.Task, signal chan struct{}) {
		for {
			select {
			case <-signal:
				close(tasks)
				return
			default:
				p.ProcessWebsites(tasks)
			}
		}
	}(proc, workerChan, signal)
	defer close(signal)

	websiteHandler := handler.NewWebsite(websiteRepo)
	newsHandler := handler.NewNews(newsRepo)

	http.HandleFunc("/websites", websiteHandler.HandlePostWebsite)
	http.HandleFunc("/news", newsHandler.HandleGetNews)
	http.HandleFunc("/news/search", newsHandler.HandleSearchNews)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
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
