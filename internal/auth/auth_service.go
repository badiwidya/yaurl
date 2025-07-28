package auth

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"
)

func NewService(db *sql.DB, logger *slog.Logger) *service {
	return &service{
		db:     db,
		logger: logger,
	}
}

type Service interface {
	RegisterUser(context.Context, RegisterUserRequest) (*string, error)
	LoginUser(context.Context, LoginUserRequest) (*string, error)
	RemoveSession(context.Context, string) error
}

type service struct {
	db     *sql.DB
	logger *slog.Logger
}

var (
	ErrHashPassword          = errors.New("Failed to hash password")
	ErrUsernameAlreadyExists = errors.New("Username already exists in database")
	ErrInvalidCredentials    = errors.New("Incorrect username or password")
	ErrSessionNotFound       = errors.New("Session not found in database")
)

func (s *service) RegisterUser(ctx context.Context, user RegisterUserRequest) (*string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("Failed to begin register transaction", "error", err.Error())
		return nil, err
	}
	defer tx.Rollback()

	var username string
	row := tx.QueryRowContext(ctx, "SELECT username FROM users WHERE username = $1;", strings.ToLower(user.Username))

	err = row.Scan(&username)
	switch {
	case err == nil:
		return nil, ErrUsernameAlreadyExists
	case errors.Is(err, sql.ErrNoRows):
		break
	default:
		s.logger.Error("Unexpected error when scanning existing username", "error", err.Error())
		return nil, err
	}

	hashedPassword, err := hashPassword(user.Password, defaultParams)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err.Error())
		return nil, ErrHashPassword
	}

	var id int
	result := tx.QueryRowContext(
		ctx,
		"INSERT INTO users (name, username, password) VALUES ($1, $2, $3) RETURNING id;",
		user.Name,
		strings.ToLower(user.Username),
		hashedPassword,
	)
	if err := result.Scan(&id); err != nil {
		s.logger.Error("Failed to insert new user", "error", err.Error())
		return nil, err
	}

	sessionId, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO sessions (session_id, user_id, expires_at) VALUES ($1, $2, $3);",
		sessionId,
		id,
		sessionMaxAge,
	)
	if err != nil {
		s.logger.Error("Failed to insert new session", "error", err.Error())
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("Failed to commit registration transaction", "error", err.Error())
		return nil, err
	}

	s.logger.Info("New user created", "username", strings.ToLower(user.Username))

	return &sessionId, nil
}

func (s *service) LoginUser(ctx context.Context, user LoginUserRequest) (*string, error) {
	var id int
	var password string
	row := s.db.QueryRowContext(ctx, "SELECT id, password FROM users WHERE username = $1;", strings.ToLower(user.Username))

	err := row.Scan(&id, &password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		s.logger.Error("Unexpected error when scanning row", "error", err.Error())
		return nil, err
	}

	isMatch, err := comparePassAndHash(user.Password, password)
	if err != nil {
		return nil, err
	}

	if !isMatch {
		return nil, ErrInvalidCredentials
	}

	sessionId, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	_, err = s.db.ExecContext(
		ctx,
		"INSERT INTO sessions (session_id, user_id, expires_at) VALUES ($1, $2, $3);",
		sessionId,
		id,
		sessionMaxAge,
	)
	if err != nil {
		s.logger.Error("Unexpected error when executing query", "error", err.Error())
		return nil, err
	}

	return &sessionId, nil
}

func (s *service) RemoveSession(ctx context.Context, sessionId string) error {
	result, err := s.db.ExecContext(
		ctx,
		"DELETE FROM sessions WHERE session_id = $1;",
		sessionId,
	)
	if err != nil {
		s.logger.Error("Unexpected error when executing query", "error", err.Error())
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return ErrSessionNotFound
	}

	return nil
}
