package shortener

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/badiwidya/yaurl/internal/pkg/middlewares"
	"github.com/badiwidya/yaurl/internal/pkg/utils"
)

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

type Handler interface {
	ShortenURL(w http.ResponseWriter, r *http.Request)
	RedirectUrl(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	service Service
}

func (h *handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	contextValue := r.Context().Value(middlewares.UserKey)

	userId, ok := contextValue.(int)
	if !ok {
		utils.JSONResponse(w, http.StatusUnauthorized, &utils.Response{
			Message: "Unauthorized",
		})
		return
	}

	ctx, close := context.WithTimeout(r.Context(), 5*time.Second)
	defer close()

	var longUrl URL
	if err := utils.ParseJSON(w, r, &longUrl); err != nil {
		var mr *utils.MalformedRequest
		if errors.As(err, &mr) {
			utils.JSONResponse(w, mr.Code, &utils.Response{
				Message: mr.Message,
			})
			return
		}
		utils.JSONResponse(w, http.StatusInternalServerError, &utils.Response{
			Message: "Internal Server Error",
		})
		return
	}

	newURL, err := h.service.CreateNewShortUrl(ctx, longUrl.Url, userId, longUrl.Expires)
	if err != nil {
		if err == ErrNotValidUrl {
			utils.JSONResponse(w, http.StatusBadRequest, &utils.Response{
				Message: "Invalid URL",
			})
			return
		}
		utils.JSONResponse(w, http.StatusInternalServerError, &utils.Response{
			Message: "Internal Server Error",
		})
		return
	}

	utils.JSONResponse(w, http.StatusCreated, &utils.Response{
		Message: "Short URL created",
		Data: map[string]string{
			"url": *newURL,
		},
	})
}

func (h *handler) RedirectUrl(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	ctx, close := context.WithTimeout(r.Context(), 5*time.Second)
	defer close()

	long_url, err := h.service.FindLongUrl(ctx, code)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, *long_url, http.StatusFound)
}
