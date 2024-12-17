package endpoint

import (
	"bytes"
	_ "embed"
	"html/template"
	"net/http"

	"github.com/MunifTanjim/stremthru/internal/config"
)

//go:embed root.html
var templateBlob string

type rootTemplateDataAddon struct {
	Name string
	URL  string
}

type RootTemplateData struct {
	Title   string
	Version string
	Addons  []rootTemplateDataAddon
}

var ExecuteTemplate = func() func(data *RootTemplateData) (bytes.Buffer, error) {
	tmpl := template.Must(template.New("root.html").Parse(templateBlob))
	return func(data *RootTemplateData) (bytes.Buffer, error) {
		var buf bytes.Buffer
		err := tmpl.Execute(&buf, data)
		return buf, err
	}
}()

func handleRoot(w http.ResponseWriter, r *http.Request) {
	td := &RootTemplateData{
		Title:   "StremThru",
		Version: config.Version,
		Addons:  []rootTemplateDataAddon{},
	}
	if config.StremioAddon.IsEnabled("store") {
		td.Addons = append(td.Addons, rootTemplateDataAddon{
			Name: "Store",
			URL:  "/stremio/store/configure",
		})
	}

	buf, err := ExecuteTemplate(td)
	if err != nil {
		SendError(w, err)
		return
	}
	SendHTML(w, 200, buf)
}

func AddRootEndpoint(mux *http.ServeMux) {
	mux.HandleFunc("/", handleRoot)
}
