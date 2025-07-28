package shortener

import "time"

type URL struct {
	Url     string     `json:"url"`
	Expires *time.Time `json:"expires,omitempty"`
}
