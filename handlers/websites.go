package handlers

import (
	"errors"
	"github.com/IlyushaZ/parser/models"
	"github.com/IlyushaZ/parser/storage"
	"github.com/mailru/easyjson"
	"net/http"
	"net/url"
)

var (
	errEmptyField = errors.New("you have to fill all the fields")
	errInvalidURL = errors.New("given url is invalid")
)

//easyjson:json
type body struct {
	MainURL      string `json:"main_url"`
	URLPattern   string `json:"url_pattern"`
	TitlePattern string `json:"title_pattern"`
	TextPattern  string `json:"text_pattern"`
}

type WebsiteHandler struct {
	repo storage.WebsiteRepository
}

func NewWebsiteHandler(repo storage.WebsiteRepository) WebsiteHandler {
	return WebsiteHandler{repo: repo}
}

func (h WebsiteHandler) HandlePostWebsite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var reqBody body
	if err := easyjson.UnmarshalFromReader(r.Body, &reqBody); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if err := validateRequest(reqBody); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	model := models.NewWebsite(reqBody.MainURL, reqBody.URLPattern, reqBody.TitlePattern, reqBody.TextPattern)
	if err := h.repo.Insert(model); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	return
}

func validateRequest(body body) error {
	if body.MainURL == "" ||
		body.URLPattern == "" ||
		body.TitlePattern == "" ||
		body.TextPattern == "" {
		return errEmptyField
	}

	if _, err := url.Parse(body.MainURL); err != nil {
		return errInvalidURL
	}

	return nil
}
