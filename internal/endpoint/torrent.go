package endpoint

import (
	"net/http"
	"strings"

	"github.com/MunifTanjim/stremthru/internal/buddy"
	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/MunifTanjim/stremthru/internal/peer_token"
	"github.com/MunifTanjim/stremthru/internal/server"
	"github.com/MunifTanjim/stremthru/internal/shared"
	"github.com/MunifTanjim/stremthru/internal/torrent_info"
)

type RecordTorrentsPayload struct {
	Items []torrent_info.TorrentItem `json:"items"`
}

func handleRecordTorrents(w http.ResponseWriter, r *http.Request) {
	peerToken := r.Header.Get("X-StremThru-Peer-Token")
	isValidToken, err := peer_token.IsValid(peerToken)
	if err != nil {
		SendError(w, r, err)
		return
	}
	if !isValidToken {
		shared.ErrorUnauthorized(r).Send(w, r)
		return
	}

	payload := &RecordTorrentsPayload{}
	if err := shared.ReadRequestBodyJSON(r, payload); err != nil {
		SendError(w, r, err)
		return
	}

	go torrent_info.Upsert(payload.Items, "", false)
	w.WriteHeader(204)
}

func handleListTorrents(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	sid := query.Get("sid")
	if sid == "" {
		shared.ErrorBadRequest(r, "missing sid").Send(w, r)
		return
	}
	if !strings.HasPrefix(sid, "tt") {
		shared.ErrorBadRequest(r, "unsupported sid").Send(w, r)
		return
	}

	originInstanceId := r.Header.Get(server.HEADER_ORIGIN_INSTANCE_ID)
	if originInstanceId == "" {
		w.Header().Set(server.HEADER_ORIGIN_INSTANCE_ID, originInstanceId)
	} else {
		w.Header().Set(server.HEADER_ORIGIN_INSTANCE_ID, config.InstanceId)

	}

	data, err := buddy.ListTorrentsByStremId(sid, query.Get("local_only") != "", originInstanceId)
	SendResponse(w, r, 200, data, err)
}

func handleTorrents(w http.ResponseWriter, r *http.Request) {
	if shared.IsMethod(r, http.MethodPost) {
		handleRecordTorrents(w, r)
		return
	}
	if shared.IsMethod(r, http.MethodGet) {
		handleListTorrents(w, r)
		return
	}
	shared.ErrorMethodNotAllowed(r).Send(w, r)
}

func AddTorrentEndpoints(mux *http.ServeMux) {
	mux.HandleFunc("/v0/torrents", handleTorrents)
}
