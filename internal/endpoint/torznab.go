package endpoint

import (
	"net/http"

	"github.com/MunifTanjim/stremthru/internal/shared"
	"github.com/MunifTanjim/stremthru/internal/torznab"
)

func handleTorznab(w http.ResponseWriter, r *http.Request) {
	t := r.URL.Query().Get("t")

	if t == "" {
		http.Redirect(w, r, r.URL.Path+"?t=caps", http.StatusTemporaryRedirect)
		return
	}

	switch t {
	case "caps":
		shared.SendXML(w, r, 200, torznab.StremThruIndexer.Capabilities())
	case "tvsearch", "movie":
		query, err := torznab.ParseQuery(r.URL.Query())
		if err != nil {
			shared.SendXML(w, r, 200, torznab.ErrorIncorrectParameter(err.Error()))
			return
		}
		if query.IMDBId == "" {
			shared.SendXML(w, r, 200, torznab.ErrorMissingParameter("imdbid"))
			return
		}
		items, err := torznab.StremThruIndexer.Search(query)
		if err != nil {
			shared.SendXML(w, r, 200, torznab.ErrorUnknownError(err.Error()))
			return
		}
		shared.SendXML(w, r, 200, torznab.ResultFeed{
			Info:  torznab.StremThruIndexer.Info(),
			Items: items,
		})
	default:
		shared.SendXML(w, r, 200, torznab.ErrorIncorrectParameter(t))
	}
}
func AddTorznabEndpoints(mux *http.ServeMux) {
	mux.HandleFunc("/v0/torznab/api", handleTorznab)
}
