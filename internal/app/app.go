package app

import (
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/badiwidya/yaurl/internal/config"
	authHandler "github.com/badiwidya/yaurl/internal/handler/auth"
	shortenerHandler "github.com/badiwidya/yaurl/internal/handler/shortener"
	"github.com/badiwidya/yaurl/internal/middleware"
	authService "github.com/badiwidya/yaurl/internal/service/auth"
	shortenerService "github.com/badiwidya/yaurl/internal/service/shortener"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Config struct {
	APP_HOST     string
	APP_PORT     string
	APP_BASE_URL string
	DB_STRING    string
	LOG_LEVEL    string
}

type app struct {
	addr string
	cfg  *config.Config
}

func New(addr string, cfg *config.Config) *app {
	return &app{addr: addr, cfg: cfg}
}

func (a *app) Run() error {
	mux := http.NewServeMux()

	level := a.cfg.GetLogLevel()
	logger := createNewLogger(level)

	logger.Info("Database connected")
	db, err := sql.Open("pgx", a.cfg.DB_STRING)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("Cannot ping to database: %v\n", err)
	}

	urlService := shortenerService.New(a.cfg, logger.With("service", "shortener-service"), db)
	urlHandler := shortenerHandler.New(urlService)

	authS := authService.New(db, logger.With("service", "auth-service"))
	authH := authHandler.New(authS)

	shortenURL := middleware.AuthRequired(db, http.HandlerFunc(urlHandler.ShortenURL))
	logout := middleware.AuthRequired(db, http.HandlerFunc(authH.HandleLogout))

	mux.Handle("POST /api/url", shortenURL)
	mux.HandleFunc("GET /{code}", urlHandler.RedirectUrl)

	mux.HandleFunc("POST /api/auth/register", authH.HandleRegister)
	mux.HandleFunc("POST /api/auth/login", authH.HandleLogin)
	mux.Handle("POST /api/auth/logout", logout)

	logger.Info("Server started", "port", a.cfg.APP_PORT)
	return http.ListenAndServe(a.addr, mux)
}

func createNewLogger(level slog.Level) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)

	return slog.New(handler)
}
