package app

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/badiwidya/yaurl/internal/auth"
	"github.com/badiwidya/yaurl/internal/config"
	"github.com/badiwidya/yaurl/internal/pkg/middlewares"
	"github.com/badiwidya/yaurl/internal/shortener"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Server struct {
	httpServer *http.Server
	db         *sql.DB
	logger     *slog.Logger
	cfg        *config.Config
}

func NewServer(cfg *config.Config) (*Server, error) {
	level := cfg.GetLogLevel()
	logger := createNewLogger(level)

	logger.Info("Connecting to database...")
	db, err := initDatabase(cfg.DB_STRING)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err.Error())
		return nil, err
	}
	logger.Info("Database connected successfully")

	return &Server{
		db:     db,
		logger: logger,
		cfg:    cfg,
	}, nil
}

func (s *Server) Run() error {
	router := s.setupRouter()
	address := s.cfg.APP_HOST + ":" + s.cfg.APP_PORT

	s.httpServer = &http.Server{
		Addr:    address,
		Handler: router,
	}

	serverErrors := make(chan error, 1)

	go func() {
		s.logger.Info("Server starting", "address", s.httpServer.Addr)
		serverErrors <- s.httpServer.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case sig := <-quit:
		s.logger.Info("Shutdown signal received", "signal", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.logger.Info("Shutting down server...")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Server shutdown failed", "error", err.Error())
	}

	s.logger.Info("Closing database connection...")
	if err := s.db.Close(); err != nil {
		s.logger.Error("Failed to close database", "error", err.Error())
	}

	s.logger.Info("Shutdown completed")
	return nil
}

func (s *Server) setupRouter() http.Handler {
	mux := http.NewServeMux()

	shortenerService := shortener.NewService(s.cfg, s.logger.With("op", "shortener"), s.db)
	shortenerHandler := shortener.NewHandler(shortenerService)
	authService := auth.NewService(s.db, s.logger.With("op", "auth"))
	authHandler := auth.NewHandler(authService)

	authMiddleware := middlewares.NewAuthRequired(s.db)

	shortenerRoutes := shortener.RegisterRoutes(shortenerHandler, authMiddleware)
	authRoutes := auth.RegisterRoutes(authHandler, authMiddleware)

	mux.Handle("/api/auth", http.StripPrefix("/api/auth", authRoutes))
	mux.Handle("/", shortenerRoutes)

	return mux
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
		db.Close()
		return nil, err
	}

	return db, nil
}
