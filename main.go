package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

var db *sql.DB

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load env file")
	}

	slog.Info("Database connected")
	db, err = sql.Open("pgx", os.Getenv("GOOSE_DBSTRING"))
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("Cannot ping to database: %v\n", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/url", HandlePOSTURL)

	slog.Info("Server started", "port", os.Getenv("APP_PORT"))
	if err = http.ListenAndServe(
		fmt.Sprintf(
			"%s:%s",
			os.Getenv("APP_HOST"),
			os.Getenv("APP_PORT"),
		),
		mux,
	); err != nil {
		log.Fatalf("Unable to start server: %v\n", err)
	}
}

func HandlePOSTURL(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var err error

	type URL struct {
		Url string `json:"url"`
	}

	ctx, close := context.WithTimeout(r.Context(), 5*time.Second)
	defer close()

	var longUrl URL
	if err = json.NewDecoder(r.Body).Decode(&longUrl); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}

	result, err := url.Parse(longUrl.Url)
	if err != nil || result.Scheme == "" || result.Host == "" {
		http.Error(w, "Bad request: must be a valid url", http.StatusBadRequest)
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	shortCode := make([]byte, 7)
	for i := range shortCode {
		shortCode[i] = charset[seededRand.Intn(len(charset))]
	}

	_, err = db.ExecContext(
		ctx,
		"INSERT INTO urls (long_url, short_url) VALUES ($1, $2)",
		longUrl.Url,
		string(shortCode),
	)
	if err != nil {
		slog.Error("Couldn't exec insert query")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

	newURL := URL{
		Url: fmt.Sprintf("%s/%s", os.Getenv("APP_BASE_URL"), string(shortCode)),
	}
	if err = json.NewEncoder(w).Encode(newURL); err != nil {
		slog.Warn("Couldn't send json response")
	}
}
