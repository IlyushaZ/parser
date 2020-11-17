package handler

import (
	"context"
	"log"
	"net/url"

	"github.com/IlyushaZ/parser/internal/model"
	"github.com/IlyushaZ/parser/pkg/api"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errEmptyField    = errors.New("you have to fill all the fields")
	errInvalidURL    = errors.New("given url is invalid")
	errAlreadyExists = errors.New("website already exists")
)

type WebsiteRepository interface {
	Insert(*model.Website) error
	WebsiteExists(string) bool
}

func NewWebsite(repo WebsiteRepository) Website {
	return Website{repo: repo}
}

type Website struct {
	api.UnimplementedWebsiteServer
	repo WebsiteRepository
}

func (w Website) Add(ctx context.Context, req *api.AddWebsiteRequest) (*empty.Empty, error) {
	if err := w.validate(req); err != nil {
		if errors.Is(err, errAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	website := model.NewWebsite(req.GetMainUrl(), req.GetUrlPattern(), req.GetTitlePattern(), req.GetTextPattern())
	if err := w.repo.Insert(&website); err != nil {
		log.Println(err.Error())
		return nil, status.Error(codes.DataLoss, "could not save website")
	}

	return &empty.Empty{}, nil
}

func (w Website) validate(req *api.AddWebsiteRequest) error {
	if req.GetMainUrl() == "" ||
		req.GetUrlPattern() == "" ||
		req.GetTitlePattern() == "" ||
		req.GetTextPattern() == "" {
		return errEmptyField
	}

	if w.repo.WebsiteExists(req.GetMainUrl()) {
		return errAlreadyExists
	}

	if _, err := url.Parse(req.GetMainUrl()); err != nil {
		return errInvalidURL
	}

	return nil
}
