package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/badiwidya/yaurl/internal/dto"
	"github.com/badiwidya/yaurl/internal/service/auth"
)

func New(service auth.Service) *handler {
	return &handler{
		service: service,
	}
}

type Handler interface {
	RegisterHandler(http.ResponseWriter, *http.Request)
	LoginHandler(http.ResponseWriter, *http.Request)
	LogoutHandler(http.ResponseWriter, *http.Request)
}

type handler struct {
	service auth.Service
}

func (h *handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	jsonEncoder := json.NewEncoder(w)
	jsonDecoder := json.NewDecoder(r.Body)

	w.Header().Set("Content-Type", "application/json")

	var user dto.RegisterUserRequest
	if err := jsonDecoder.Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(struct{ message string }{message: "Bad request, invalid data"})
		return
	}

	if err := user.Validate(); err != nil {
		var validationErrs dto.ValidationErrors

		if errors.As(err, &validationErrs) {
			w.WriteHeader(http.StatusBadRequest)
			jsonEncoder.Encode(struct {
				message string
				errors  any
			}{
				message: "Validation error",
				errors:  validationErrs,
			})

		} else {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(struct{ message string }{message: "Internal server error"})
		}
		return
	}

	session, err := h.service.RegisterUser(ctx, user)
	if err != nil {
		if err == auth.ErrUsernameAlreadyExists {
			w.WriteHeader(http.StatusConflict)
			jsonEncoder.Encode(struct{ message string }{message: "Username already exists"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(struct{ message string }{message: "Internal server error"})
		}
		return
	}

	cookie := newSessionCookie(session)

	w.WriteHeader(http.StatusCreated)
	http.SetCookie(w, cookie)

	jsonEncoder.Encode(struct{ message string }{message: "Registered successfully"})
}

func (h *handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	jsonEncoder := json.NewEncoder(w)
	jsonDecoder := json.NewDecoder(r.Body)

	w.Header().Set("Content-Type", "application/json")

	var user dto.LoginUserRequest
	if err := jsonDecoder.Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(struct{ message string }{message: "Bad request, invalid data"})
		return
	}

	if err := user.Validate(); err != nil {
		var validationErrs dto.ValidationErrors

		if errors.As(err, &validationErrs) {
			w.WriteHeader(http.StatusBadRequest)
			jsonEncoder.Encode(struct {
				message string
				errors  any
			}{
				message: "Validation error",
				errors:  validationErrs,
			})

		} else {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(struct{ message string }{message: "Internal server error"})
		}
		return
	}

	session, err := h.service.LoginUser(ctx, user)
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			w.WriteHeader(http.StatusConflict)
			jsonEncoder.Encode(struct{ message string }{message: "Invalid credentials"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(struct{ message string }{message: "Internal server error"})
		}
		return
	}

	cookie := newSessionCookie(session)

	w.WriteHeader(http.StatusCreated)
	http.SetCookie(w, cookie)

	jsonEncoder.Encode(struct{ message string }{message: "Logged in successfully"})
}

func (h *handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()
	w.Header().Set("Content-Type", "application/json")

	jsonEncoder := json.NewEncoder(w)

	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusBadRequest)
			jsonEncoder.Encode(struct{ message string }{message: "Cookie not found"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(struct{ message string }{message: "Internal server error"})
		}
		return
	}

	if err := h.service.RemoveSession(ctx, cookie.Value); err != nil {
		if err == auth.ErrSessionNotFound {
			w.WriteHeader(http.StatusNotFound)
			jsonEncoder.Encode(struct{ message string }{message: "Session not found in database"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(struct{ message string }{message: "Internal server error"})
		}
		return
	}

	cookie.MaxAge = -1

	w.WriteHeader(http.StatusOK)
	http.SetCookie(w, cookie)
	jsonEncoder.Encode(struct{ message string }{message: "Logged out successfully"})
}

const sessionCookieName = "session_id"

func newSessionCookie(session *string) *http.Cookie {
	return &http.Cookie{
		Name:     "session_id",
		Value:    *session,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   7 * 24 * 60 * 60,
		Secure:   false, // God please remind me
		SameSite: http.SameSiteStrictMode,
	}
}
