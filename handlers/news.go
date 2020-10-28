package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/IlyushaZ/parser/models"
	"github.com/mailru/easyjson"
)

type NewsRepository interface {
	Get(limit, offset int) ([]models.News, error)
	SearchByTitle(title string) ([]models.News, error)
}

//easyjson:json
type newsArr []models.News

//easyjson:skip
type NewsHandler struct {
	repo NewsRepository
}

func NewNewsHandler(repo NewsRepository) NewsHandler {
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

	news, err := h.repo.Get(limit, offset)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	result, _ := easyjson.Marshal(newsArr(news))

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result)
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

	news, err := h.repo.SearchByTitle(query[0])
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	result, _ := easyjson.Marshal(newsArr(news))

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result)
}
