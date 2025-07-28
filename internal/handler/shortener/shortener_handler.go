package shortener

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/badiwidya/yaurl/internal/dto"
	"github.com/badiwidya/yaurl/internal/middleware"
	shortenerService "github.com/badiwidya/yaurl/internal/service/shortener"
	"github.com/badiwidya/yaurl/internal/util"
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

	contextValue := r.Context().Value(middleware.UserKey)

	userId, ok := contextValue.(int)
	if !ok {
		util.JSONResponse(w, http.StatusUnauthorized, &dto.Response{
			Message: "Unauthorized",
		})
		return
	}

	ctx, close := context.WithTimeout(r.Context(), 5*time.Second)
	defer close()

	var longUrl dto.URL
	if err := util.ParseJSON(w, r, &longUrl); err != nil {
		var mr *util.MalformedRequest
		if errors.As(err, &mr) {
			util.JSONResponse(w, mr.Code, &dto.Response{
				Message: mr.Message,
			})
			return
		}
		util.JSONResponse(w, http.StatusInternalServerError, &dto.Response{
			Message: "Internal Server Error",
		})
		return
	}

	newURL, err := s.service.CreateNewShortUrl(ctx, longUrl.Url, userId, longUrl.Expires)
	if err != nil {
		if err == shortenerService.ErrNotValidUrl {
			util.JSONResponse(w, http.StatusBadRequest, &dto.Response{
				Message: "Invalid URL",
			})
			return
		}
		util.JSONResponse(w, http.StatusInternalServerError, &dto.Response{
			Message: "Internal Server Error",
		})
		return
	}

	util.JSONResponse(w, http.StatusCreated, &dto.Response{
		Message: "Short URL created",
		Data: map[string]string{
			"url": *newURL,
		},
	})
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
