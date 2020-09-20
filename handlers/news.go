package handlers

import (
	"github.com/IlyushaZ/parser/models"
	"github.com/IlyushaZ/parser/storage"
	"github.com/mailru/easyjson"
	"net/http"
	"strconv"
)

//easyjson:json
type newsArr []models.News

//easyjson:json
type getResponseBody struct {
	News newsArr `json:"news"`
}

type NewsHandler struct {
	repo storage.NewsRepository
}

func NewNewsHandler(repo storage.NewsRepository) NewsHandler {
	return NewsHandler{repo: repo}
}

func (h NewsHandler) HandleGetNews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	limit := 10
	offset := 0
	var err error

	param, ok := r.URL.Query()["limit"]
	if ok && len(param) > 0 {
		limit, err = strconv.Atoi(param[0])
	}

	param, ok = r.URL.Query()["offset"]
	if ok && len(param) > 0 {
		offset, err = strconv.Atoi(param[0])
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := easyjson.Marshal(newsArr(h.repo.Get(limit, offset)))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(result)
}

func (h NewsHandler) HandleSearchNews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	query, ok := r.URL.Query()["q"]
	if !ok || len(query) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := easyjson.Marshal(newsArr(h.repo.SearchByTitle(query[0])))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(result)
}
