package shortener

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

func NewService(config *config.Config, logger *slog.Logger, db *sql.DB) *service {
	return &service{
		cfg:    config,
		logger: logger,
		db:     db,
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

type Service interface {
	CreateNewShortUrl(context.Context, string, int, *time.Time) (*string, error)
	FindLongUrl(context.Context, string) (*string, error)
}

type service struct {
	cfg    *config.Config
	logger *slog.Logger
	db     *sql.DB
	rand   *rand.Rand
}

var ErrNotValidUrl error = errors.New("Invalid URL")
var ErrExecQuery error = errors.New("Error when executing query")
var ErrNotFound error = errors.New("Row not found")

func (s *service) FindLongUrl(ctx context.Context, code string) (*string, error) {
	row := s.db.QueryRowContext(
		ctx,
		"SELECT long_url FROM urls WHERE short_url = $1",
		code,
	)

	var long_url string

	err := row.Scan(&long_url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		s.logger.Error("An error occurred when scanning row", "error", err.Error())
		return nil, ErrExecQuery
	}

	return &long_url, nil
}

func (s *service) CreateNewShortUrl(ctx context.Context, longUrl string, userId int, expire *time.Time) (*string, error) {

	result, err := url.Parse(longUrl)
	if err != nil || result.Scheme == "" || result.Host == "" {
		return nil, ErrNotValidUrl
	}

	shortCode := s.generateRandomCode()

	if expire != nil {
		_, err = s.db.ExecContext(
			ctx,
			"INSERT INTO urls (user_id, long_url, short_url, expires_at) VALUES ($1, $2, $3, $4)",
			userId,
			longUrl,
			shortCode,
			expire,
		)
	} else {
		_, err = s.db.ExecContext(
			ctx,
			"INSERT INTO urls (user_id, long_url, short_url) VALUES ($1, $2, $3)",
			userId,
			longUrl,
			shortCode,
		)
	}
	if err != nil {
		s.logger.Error("Failed to execute insert query", "error", err.Error())
		return nil, ErrExecQuery
	}

	newURL := s.cfg.APP_BASE_URL + "/" + shortCode

	return &newURL, nil
}

func (s *service) generateRandomCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	shortCode := make([]byte, 7)
	for i := range shortCode {
		shortCode[i] = charset[s.rand.Intn(len(charset))]
	}

	return string(shortCode)
}
