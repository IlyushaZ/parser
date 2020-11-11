package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
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
		log.Println(err)
		return
	}
	defer pool.Close()

	websiteRepo := storage.NewWebsiteRepository(pool)
	newsRepo := storage.NewNewsRepository(pool)

	mc := memcache.New(mcURL)
	if err = mc.Ping(); err != nil {
		log.Println(err)
		return
	}
	mc.MaxIdleConns = 10
	if err = mc.DeleteAll(); err != nil {
		panic(err)
	}

	newsCache := storage.NewNewsCache(mc, 1800)

	proc := processor.New(websiteRepo, newsRepo, newsCache)

	var wg sync.WaitGroup
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go proc.ProcessNews(&wg)
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	proc.Process(ctx, &wg)

	websiteHandler := handler.NewWebsite(websiteRepo)
	newsHandler := handler.NewNews(newsRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("/websites", websiteHandler.HandlePostWebsite)
	mux.HandleFunc("/news", newsHandler.HandleGetNews)
	mux.HandleFunc("/news/search", newsHandler.HandleSearchNews)

	srv := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Println(err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	<-sigCh

	if err := srv.Shutdown(context.Background()); err != nil {
		panic(err)
	}
	cancel()
	wg.Wait()
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
