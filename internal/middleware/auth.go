package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
)

type userKey string

const UserKey userKey = "userIdKey"

func AuthRequired(db *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userId int

		cookie, err := r.Cookie("session_id")
		if err != nil {
			if err == http.ErrNoCookie {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(struct{ message string }{message: "Unauthorized"})
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(struct{ message string }{message: "Internal server error"})
			}
			return
		}

		row := db.QueryRowContext(r.Context(), "SELECT user_id FROM sessions where session_id = $1", cookie.Value)

		if err := row.Scan(&userId); err != nil {
			if err == sql.ErrNoRows {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(struct{ message string }{message: "Unauthorized"})
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(struct{ message string }{message: "Internal server error"})
			}
			return
		}

		ctx := context.WithValue(r.Context(), UserKey, userId)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
