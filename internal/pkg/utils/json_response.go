package utils

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
)

type Response struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func JSONResponse(w http.ResponseWriter, status int, response *Response) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(*response)
}

type MalformedRequest struct {
	Code    int
	Message string
}

func (m MalformedRequest) Error() string {
	return m.Message
}

func ParseJSON(w http.ResponseWriter, r *http.Request, dest any) error {
	ct := r.Header.Get("Content-Type")
	if ct != "" {
		mediatype, _, err := mime.ParseMediaType(ct)
		if err != nil || mediatype != "application/json" {
			return &MalformedRequest{
				Code:    http.StatusUnsupportedMediaType,
				Message: "Content-Type is unsupported, expected application/json",
			}
		}
	}

	jsonDecoder := json.NewDecoder(r.Body)

	err := jsonDecoder.Decode(&dest)
	if err != nil {
		var syntaxError *json.SyntaxError

		switch {
		case errors.As(err, &syntaxError):
			return &MalformedRequest{
				Code:    http.StatusBadRequest,
				Message: "Request body contains badly-formed JSON",
			}
		case errors.Is(err, io.EOF):
			return &MalformedRequest{
				Code:    http.StatusBadRequest,
				Message: "Request body must not be empty",
			}
		default:
			return err
		}
	}

	return nil
}
