package stremio_sidekick

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/MunifTanjim/stremthru/internal/shared"
)

var IsMethod = shared.IsMethod
var SendError = shared.SendError
var ExtractRequestBaseURL = shared.ExtractRequestBaseURL

func SendResponse(w http.ResponseWriter, r *http.Request, statusCode int, data any) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(data); err != nil {
		LogError(r, "failed to encode json", err)
	}
}

func SendHTML(w http.ResponseWriter, statusCode int, data bytes.Buffer) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	shared.SendHTML(w, statusCode, data)
}
