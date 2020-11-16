package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/IlyushaZ/parser/internal/handler"
	"github.com/IlyushaZ/parser/internal/processor"
	"github.com/IlyushaZ/parser/internal/storage"
	"github.com/IlyushaZ/parser/pkg/api"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
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
		log.Println(err.Error())
		return
	}
	defer pool.Close()

	websiteRepo := storage.NewWebsiteRepository(pool)
	newsRepo := storage.NewNewsRepository(pool)

	mc := memcache.New(mcURL)
	if err = mc.Ping(); err != nil {
		log.Println(err.Error())
		return
	}
	mc.MaxIdleConns = 10
	if err = mc.DeleteAll(); err != nil {
		log.Println(err.Error())
		return
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

	srv := grpc.NewServer()
	api.RegisterWebsiteServer(srv, websiteHandler)
	api.RegisterNewsServer(srv, newsHandler)

	l, err := net.Listen("tcp", ":8080") //nolint:gosec
	if err != nil {
		log.Println(err)

		cancel()
		wg.Wait()
		return
	}

	go func() {
		if err := srv.Serve(l); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			cancel()
			wg.Wait()

			panic(err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	<-sigCh

	srv.Stop()
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
