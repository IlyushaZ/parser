package handler

import (
	"log"
	"net/http"
	"net/url"

	"github.com/IlyushaZ/parser/internal/model"
	"github.com/mailru/easyjson"
	"github.com/pkg/errors"
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

//easyjson:json
type body struct {
	MainURL      string `json:"main_url"`
	URLPattern   string `json:"url_pattern"`
	TitlePattern string `json:"title_pattern"`
	TextPattern  string `json:"text_pattern"`
}

//easyjson:skip
type Website struct {
	repo WebsiteRepository
}

func NewWebsite(repo WebsiteRepository) Website {
	return Website{repo: repo}
}

func (wh Website) HandlePostWebsite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var reqBody body
	if err := easyjson.UnmarshalFromReader(r.Body, &reqBody); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if err := wh.validateRequest(reqBody); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	website := model.NewWebsite(reqBody.MainURL, reqBody.URLPattern, reqBody.TitlePattern, reqBody.TextPattern)
	if err := wh.repo.Insert(&website); err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (wh Website) validateRequest(body body) error {
	if body.MainURL == "" ||
		body.URLPattern == "" ||
		body.TitlePattern == "" ||
		body.TextPattern == "" {
		return errEmptyField
	}

	if wh.repo.WebsiteExists(body.MainURL) {
		return errAlreadyExists
	}

	if _, err := url.Parse(body.MainURL); err != nil {
		return errInvalidURL
	}

	return nil
}
