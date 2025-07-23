package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"math/rand"
	"net/url"
	"time"

	"github.com/badiwidya/yaurl/internal/config"
)

func New(config *config.Config, logger *slog.Logger, db *sql.DB) *shortenerService {
	return &shortenerService{
		cfg:    config,
		logger: logger,
		db:     db,
	}
}

type ShortenerService interface {
	CreateNewShortUrl(context.Context, string) (*string, error)
}

type shortenerService struct {
	cfg    *config.Config
	logger *slog.Logger
	db     *sql.DB
}

var ErrNotValidUrl error = errors.New("Invalid URL")
var ErrExecQuery error = errors.New("Error when executing query")

func (s *shortenerService) CreateNewShortUrl(ctx context.Context, longUrl string) (*string, error) {

	result, err := url.Parse(longUrl)
	if err != nil || result.Scheme == "" || result.Host == "" {
		return nil, ErrNotValidUrl
	}

	shortCode := s.generateRandomCode()

	_, err = s.db.ExecContext(
		ctx,
		"INSERT INTO urls (long_url, short_url) VALUES ($1, $2)",
		longUrl,
		shortCode,
	)
	if err != nil {
		s.logger.Error("Failed to execute insert query", "error", err.Error())
		return nil, ErrExecQuery
	}

	newURL := s.cfg.APP_BASE_URL + "/" + shortCode

	return &newURL, nil
}

func (s *shortenerService) generateRandomCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	shortCode := make([]byte, 7)
	for i := range shortCode {
		shortCode[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(shortCode)
}
