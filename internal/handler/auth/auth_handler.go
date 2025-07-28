package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/badiwidya/yaurl/internal/dto"
	"github.com/badiwidya/yaurl/internal/service/auth"
	"github.com/badiwidya/yaurl/internal/util"
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

	var user dto.RegisterUserRequest
	if err := util.ParseJSON(w, r, &user); err != nil {
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

	if err := user.Validate(); err != nil {
		var validationErrs dto.ValidationErrors

		if errors.As(err, &validationErrs) {
			util.JSONResponse(w, http.StatusBadRequest, &dto.Response{
				Message: "Validation error",
				Data:    validationErrs,
			})
			return
		}
		util.JSONResponse(w, http.StatusInternalServerError, &dto.Response{
			Message: "Internal Server Error",
		})
		return
	}

	session, err := h.service.RegisterUser(ctx, user)
	if err != nil {
		if err == auth.ErrUsernameAlreadyExists {
			util.JSONResponse(w, http.StatusConflict, &dto.Response{
				Message: "Username already exists",
			})
		}
		util.JSONResponse(w, http.StatusInternalServerError, &dto.Response{
			Message: "Internal Server Error",
		})
		return
	}

	cookie := newSessionCookie(session)

	util.JSONResponse(w, http.StatusCreated, &dto.Response{
		Message: "User registered successfully",
	})
	http.SetCookie(w, cookie)
}

func (h *handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	var user dto.LoginUserRequest
	if err := util.ParseJSON(w, r, &user); err != nil {
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

	if err := user.Validate(); err != nil {
		var validationErrs dto.ValidationErrors

		if errors.As(err, &validationErrs) {
			util.JSONResponse(w, http.StatusBadRequest, &dto.Response{
				Message: "Validation error",
				Data:    validationErrs,
			})
			return
		}
		util.JSONResponse(w, http.StatusInternalServerError, &dto.Response{
			Message: "Internal Server Error",
		})
		return
	}

	session, err := h.service.LoginUser(ctx, user)
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			util.JSONResponse(w, http.StatusUnauthorized, &dto.Response{
				Message: "Invalid credentials",
			})
			return
		}
		util.JSONResponse(w, http.StatusInternalServerError, &dto.Response{
			Message: "Internal Server Error",
		})
		return
	}

	cookie := newSessionCookie(session)

	http.SetCookie(w, cookie)
	util.JSONResponse(w, http.StatusOK, &dto.Response{
		Message: "User logged in successfully",
	})
}

func (h *handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			util.JSONResponse(w, http.StatusUnauthorized, &dto.Response{
				Message: "Unauthorized",
			})
			return
		}
		util.JSONResponse(w, http.StatusInternalServerError, &dto.Response{
			Message: "Internal Server Error",
		})
		return
	}

	if err := h.service.RemoveSession(ctx, cookie.Value); err != nil {
		if err == auth.ErrSessionNotFound {
			util.JSONResponse(w, http.StatusUnauthorized, &dto.Response{
				Message: "Unauthorized",
			})
			return
		}
		util.JSONResponse(w, http.StatusInternalServerError, &dto.Response{
			Message: "Internal Server Error",
		})
		return
	}

	cookie.MaxAge = -1

	http.SetCookie(w, cookie)
	util.JSONResponse(w, http.StatusOK, &dto.Response{
		Message: "User logged out successfully",
	})
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
