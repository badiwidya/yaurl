package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/badiwidya/yaurl/internal/pkg/types"
	"github.com/badiwidya/yaurl/internal/pkg/utils"
)

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

type Handler interface {
	HandleRegister(http.ResponseWriter, *http.Request)
	HandleLogin(http.ResponseWriter, *http.Request)
	HandleLogout(http.ResponseWriter, *http.Request)
}

type handler struct {
	service Service
}

func (h *handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	var user RegisterUserRequest
	if err := utils.ParseJSON(w, r, &user); err != nil {
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

	if err := user.Validate(); err != nil {
		var validationErrs types.ValidationErrors

		if errors.As(err, &validationErrs) {
			utils.JSONResponse(w, http.StatusBadRequest, &utils.Response{
				Message: "Validation error",
				Data:    validationErrs,
			})
			return
		}
		utils.JSONResponse(w, http.StatusInternalServerError, &utils.Response{
			Message: "Internal Server Error",
		})
		return
	}

	session, err := h.service.RegisterUser(ctx, user)
	if err != nil {
		if err == ErrUsernameAlreadyExists {
			utils.JSONResponse(w, http.StatusConflict, &utils.Response{
				Message: "Username already exists",
			})
		}
		utils.JSONResponse(w, http.StatusInternalServerError, &utils.Response{
			Message: "Internal Server Error",
		})
		return
	}

	cookie := newSessionCookie(session)

	utils.JSONResponse(w, http.StatusCreated, &utils.Response{
		Message: "User registered successfully",
	})
	http.SetCookie(w, cookie)
}

func (h *handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	var user LoginUserRequest
	if err := utils.ParseJSON(w, r, &user); err != nil {
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

	if err := user.Validate(); err != nil {
		var validationErrs types.ValidationErrors

		if errors.As(err, &validationErrs) {
			utils.JSONResponse(w, http.StatusBadRequest, &utils.Response{
				Message: "Validation error",
				Data:    validationErrs,
			})
			return
		}
		utils.JSONResponse(w, http.StatusInternalServerError, &utils.Response{
			Message: "Internal Server Error",
		})
		return
	}

	session, err := h.service.LoginUser(ctx, user)
	if err != nil {
		if err == ErrInvalidCredentials {
			utils.JSONResponse(w, http.StatusUnauthorized, &utils.Response{
				Message: "Invalid credentials",
			})
			return
		}
		utils.JSONResponse(w, http.StatusInternalServerError, &utils.Response{
			Message: "Internal Server Error",
		})
		return
	}

	cookie := newSessionCookie(session)

	http.SetCookie(w, cookie)
	utils.JSONResponse(w, http.StatusOK, &utils.Response{
		Message: "User logged in successfully",
	})
}

func (h *handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			utils.JSONResponse(w, http.StatusUnauthorized, &utils.Response{
				Message: "Unauthorized",
			})
			return
		}
		utils.JSONResponse(w, http.StatusInternalServerError, &utils.Response{
			Message: "Internal Server Error",
		})
		return
	}

	if err := h.service.RemoveSession(ctx, cookie.Value); err != nil {
		if err == ErrSessionNotFound {
			utils.JSONResponse(w, http.StatusUnauthorized, &utils.Response{
				Message: "Unauthorized",
			})
			return
		}
		utils.JSONResponse(w, http.StatusInternalServerError, &utils.Response{
			Message: "Internal Server Error",
		})
		return
	}

	cookie.MaxAge = -1

	http.SetCookie(w, cookie)
	utils.JSONResponse(w, http.StatusOK, &utils.Response{
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
