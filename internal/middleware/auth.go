package middleware

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/badiwidya/yaurl/internal/dto"
	"github.com/badiwidya/yaurl/internal/util"
)

type userKey string

const UserKey userKey = "userIdKey"

func AuthRequired(db *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userId int

		cookie, err := r.Cookie("session_id")
		if err != nil {
			if err == http.ErrNoCookie {
				util.JSONResponse(w, http.StatusUnauthorized, &dto.Response{
					Message: "Unauthorized",
				})
			} else {
				util.JSONResponse(w, http.StatusBadRequest, &dto.Response{
					Message: "Internal server error",
				})
			}
			return
		}

		row := db.QueryRowContext(r.Context(), "SELECT user_id FROM sessions where session_id = $1", cookie.Value)

		if err := row.Scan(&userId); err != nil {
			if err == sql.ErrNoRows {
				util.JSONResponse(w, http.StatusUnauthorized, &dto.Response{
					Message: "Unauthorized",
				})
			} else {
				util.JSONResponse(w, http.StatusBadRequest, &dto.Response{
					Message: "Internal server error",
				})
			}
			return
		}

		ctx := context.WithValue(r.Context(), UserKey, userId)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
