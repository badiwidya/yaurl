package util

import (
	"encoding/json"
	"net/http"

	"github.com/badiwidya/yaurl/internal/dto"
)

func JSONResponse(w http.ResponseWriter, status int, response *dto.Response) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(*response)
}
