package middlewares

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/badiwidya/yaurl/internal/pkg/utils"
)

type userKey string

const UserKey userKey = "userIdKey"

func NewAuthRequired(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userId int
			cookie, err := r.Cookie("session_id")
			if err != nil {
				if errors.Is(err, http.ErrNoCookie) {
					utils.JSONResponse(w, http.StatusUnauthorized, &utils.Response{
						Message: "Unauthorized",
					})
				} else {
					utils.JSONResponse(w, http.StatusBadRequest, &utils.Response{
						Message: "Internal server error",
					})
				}
				return
			}

			row := db.QueryRowContext(r.Context(),
				"SELECT user_id FROM sessions where session_id = $1 AND expires_at > NOW()",
				cookie.Value,
			)

			if err := row.Scan(&userId); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					utils.JSONResponse(w, http.StatusUnauthorized, &utils.Response{
						Message: "Unauthorized",
					})
				} else {
					utils.JSONResponse(w, http.StatusBadRequest, &utils.Response{
						Message: "Internal server error",
					})
				}
				return
			}

			ctx := context.WithValue(r.Context(), UserKey, userId)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
