package auth

import "net/http"

func RegisterRoutes(handler Handler, middleware func(http.Handler) http.Handler) *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("POST /register", handler.HandleRegister)
	r.HandleFunc("POST /login", handler.HandleLogin)
	r.Handle("POST /logout", middleware(http.HandlerFunc(handler.HandleLogout)))

	return r
}
