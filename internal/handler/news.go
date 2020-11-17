package handler

import (
	"context"
	"errors"
	"log"

	"github.com/IlyushaZ/parser/internal/model"
	"github.com/IlyushaZ/parser/internal/storage"
	"github.com/IlyushaZ/parser/pkg/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NewsRepository interface {
	Get(limit, offset int) ([]model.News, error)
	SearchByTitle(title string) ([]model.News, error)
}

type News struct {
	api.UnimplementedNewsServer
	repo NewsRepository
}

func NewNews(repo NewsRepository) News {
	return News{repo: repo}
}

func (n News) Get(ctx context.Context, req *api.GetNewsRequest) (*api.NewsResponse, error) {
	if req.GetLimit() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "limit should be more than 0")
	}

	if req.GetOffset() < 0 {
		return nil, status.Error(codes.InvalidArgument, "offset cannot be less than 0")
	}

	news, err := n.repo.Get(int(req.GetLimit()), int(req.GetOffset()))
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		log.Println(err.Error())
		return nil, status.Error(codes.Internal, "")
	}

	resp := &api.NewsResponse{}
	resp.News = make([]*api.NewsResponse_News, 0, req.GetLimit())
	for i := range news {
		respElem := api.NewsResponse_News{}

		respElem.Id = news[i].ID
		respElem.Url = news[i].URL
		respElem.Title = news[i].Title
		respElem.Text = news[i].Text

		resp.News = append(resp.News, &respElem)
	}

	return resp, nil
}

func (n News) Search(ctx context.Context, req *api.SearchNewsRequest) (*api.NewsResponse, error) {
	if req.GetQuery() == "" {
		return nil, status.Error(codes.InvalidArgument, "search query is not entered")
	}

	news, err := n.repo.SearchByTitle(req.GetQuery())
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		log.Println(err)
		return nil, status.Error(codes.Internal, "")
	}

	resp := &api.NewsResponse{}
	resp.News = make([]*api.NewsResponse_News, 0, len(news))
	for i := range news {
		respElem := api.NewsResponse_News{}

		respElem.Id = news[i].ID
		respElem.Url = news[i].URL
		respElem.Title = news[i].Title
		respElem.Text = news[i].Text

		resp.News = append(resp.News, &respElem)
	}

	return resp, nil
}
