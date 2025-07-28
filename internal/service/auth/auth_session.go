package auth

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

var (
	sessionSaltLength = 32
	sessionMaxAge     = time.Now().UTC().Add(7 * 24 * time.Hour)
)

func generateSessionID() (string, error) {
	b := make([]byte, sessionSaltLength)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
