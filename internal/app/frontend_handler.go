package app

import (
	"html/template"
	"net/http"
	"path/filepath"
)

func (s *Server) handleHomepage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			IsAuthenticated bool
		}{
			IsAuthenticated: false,
		}

		cookie, err := r.Cookie("session_id")
		if err == nil && cookie.Value != "" {
			var userId int

			row := s.db.QueryRowContext(r.Context(),
				"SELECT user_id FROM sessions where session_id = $1 AND expires_at > NOW()",
				cookie.Value,
			)

			if err := row.Scan(&userId); err != nil {
				data.IsAuthenticated = false
			} else {
				data.IsAuthenticated = true
			}
		}

		s.serveTemplate(w, "index.gohtml", data)
	}
}

func (s *Server) serveTemplate(w http.ResponseWriter, pageName string, data any) {
	pagePath := filepath.Join("templates", pageName)
	layoutPath := filepath.Join("templates", "layout.gohtml")

	tmpl, err := template.ParseFiles(layoutPath, pagePath)
	if err != nil {
		s.logger.Error("Failed to parse html templates", "error", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		s.logger.Error("Failed to execute template", "error", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
