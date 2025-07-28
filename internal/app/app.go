package app

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	"github.com/badiwidya/yaurl/internal/auth"
	"github.com/badiwidya/yaurl/internal/config"
	"github.com/badiwidya/yaurl/internal/pkg/middlewares"
	"github.com/badiwidya/yaurl/internal/shortener"
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
	db, err := initDatabase(a.cfg.DB_STRING)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err.Error())
		os.Exit(1)
	}

	shortenerService := shortener.NewService(a.cfg, logger.With("op", "shortener"), db)
	shortenerHandler := shortener.NewHandler(shortenerService)
	authService := auth.NewService(db, logger.With("op", "auth"))
	authHandler := auth.NewHandler(authService)

	authMiddleware := middlewares.NewAuthRequired(db)

	shortenerRoutes := shortener.RegisterRoutes(shortenerHandler, authMiddleware)
	authRoutes := auth.RegisterRoutes(authHandler, authMiddleware)

	mux.Handle("/api/auth", http.StripPrefix("/api/auth", authRoutes))
	mux.Handle("/", shortenerRoutes)

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

func initDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
