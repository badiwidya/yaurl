package app

import (
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/badiwidya/yaurl/internal/config"
	shortenerHandler "github.com/badiwidya/yaurl/internal/handler/shortener"
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

	service := shortenerService.New(a.cfg, logger.With("service", "shortener-service"), db)

	handler := shortenerHandler.New(service)

	mux.HandleFunc("/api/url", handler.ShortenURL)
	mux.HandleFunc("/{code}", handler.RedirectUrl)

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
