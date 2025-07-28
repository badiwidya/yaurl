package shortener

import "net/http"

func RegisterRoutes(handler Handler, middleware func(http.Handler) http.Handler) *http.ServeMux {
	r := http.NewServeMux()

	r.Handle("POST /api/url", middleware(http.HandlerFunc(handler.ShortenURL)))
	r.HandleFunc("GET /{code}", handler.RedirectUrl)

	return r
}
