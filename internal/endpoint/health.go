package endpoint

import "net/http"

type HealthData struct {
	Status string `json:"status"`
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	health := &HealthData{}
	health.Status = "ok"
	SendResponse(w, 200, health, nil)
}

func AddHealthEndpoints(mux *http.ServeMux) {
	mux.HandleFunc("/v0/health", handleHealth)
}
