package shortener

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/badiwidya/yaurl/internal/dto"
	shortenerService "github.com/badiwidya/yaurl/internal/service/shortener"
)

func New(service shortenerService.ShortenerService) *shortenerHandler {
	return &shortenerHandler{
		service: service,
	}
}

type ShortenerHandler interface {
	ShortenURL(w http.ResponseWriter, r *http.Request)
	RedirectUrl(w http.ResponseWriter, r *http.Request)
}

type shortenerHandler struct {
	service shortenerService.ShortenerService
}

func (s *shortenerHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	ctx, close := context.WithTimeout(r.Context(), 5*time.Second)
	defer close()

	jsonEncoder := json.NewEncoder(w)

	w.Header().Set("Content-Type", "application/json")

	var longUrl dto.URL
	if err := json.NewDecoder(r.Body).Decode(&longUrl); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(struct{ message string }{message: "Not a valid URL"})
		return
	}

	newURL, err := s.service.CreateNewShortUrl(ctx, longUrl.Url)
	if err != nil {
		if err == shortenerService.ErrNotValidUrl {
			w.WriteHeader(http.StatusBadRequest)
			jsonEncoder.Encode(struct{ message string }{message: "Not a valid URL"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(struct{ message string }{message: "Internal server error"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	jsonEncoder.Encode(dto.URL{Url: *newURL})
}

func (s *shortenerHandler) RedirectUrl(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	ctx, close := context.WithTimeout(r.Context(), 5*time.Second)
	defer close()

	long_url, err := s.service.FindLongUrl(ctx, code)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, *long_url, http.StatusFound)
}
