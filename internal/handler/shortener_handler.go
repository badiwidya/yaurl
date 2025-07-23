package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/badiwidya/yaurl/internal/dto"
	"github.com/badiwidya/yaurl/internal/service"
)

func New(service service.ShortenerService) *shortenerHandler {
	return &shortenerHandler{
		service: service,
	}
}

type ShortenerHandler interface {
	ShortenURL(w http.ResponseWriter, r *http.Request)
}

type shortenerHandler struct {
	service service.ShortenerService
}

func (s *shortenerHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	ctx, close := context.WithTimeout(r.Context(), 5*time.Second)
	defer close()

	var longUrl dto.URL
	if err := json.NewDecoder(r.Body).Decode(&longUrl); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}

	jsonEncoder := json.NewEncoder(w)

	newURL, err := s.service.CreateNewShortUrl(ctx, longUrl.Url)
	if err != nil {
		if err == service.ErrNotValidUrl {
			w.WriteHeader(http.StatusBadRequest)
			jsonEncoder.Encode(struct{ message string }{message: "Bad Request: Not a valid URL"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(struct{ message string }{message: "Internal server error"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	jsonEncoder.Encode(dto.URL{Url: *newURL})
}
