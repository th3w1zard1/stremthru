package stremio_wrap

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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		LogError(r, "failed to encode json", err)
	}
}

func SendHTML(w http.ResponseWriter, statusCode int, data bytes.Buffer) {
	shared.SendHTML(w, statusCode, data)
}

func dedupeStreams(allStreams []WrappedStream) []WrappedStream {
	hashSeen := map[string]struct{}{}

	streams := []WrappedStream{}
	for i := range allStreams {
		s := allStreams[i]
		if s.r != nil && s.r.Hash != "" {
			if _, seen := hashSeen[s.r.Hash]; seen {
				continue
			} else {
				hashSeen[s.r.Hash] = struct{}{}
			}
		}
		streams = append(streams, s)
	}
	return streams
}
