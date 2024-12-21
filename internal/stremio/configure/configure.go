package configure

import (
	"bytes"
	"embed"
	"html/template"

	"github.com/MunifTanjim/stremthru/internal/config"
)

type ConfigType string

const (
	ConfigTypeText     ConfigType = "text"
	ConfigTypeNumber   ConfigType = "number"
	ConfigTypePassword ConfigType = "password"
	ConfigTypeCheckbox ConfigType = "checkbox"
	ConfigTypeSelect   ConfigType = "select"
)

type ConfigOption struct {
	Value string
	Label string
}

type Config struct {
	Key         string
	Type        ConfigType
	Default     string
	Title       string
	Description template.HTML
	Options     []ConfigOption
	Required    bool
	Error       string
}

type TemplateData struct {
	Title       string
	Description string
	Version     string
	Configs     []Config
	Error       string
	ManifestURL string
}

func (td *TemplateData) HasError() bool {
	for i := range td.Configs {
		if td.Configs[i].Error != "" {
			return true
		}
	}
	return false
}

type TemplateExecutor func(data *TemplateData, name string) (bytes.Buffer, error)
type PageGetter func(data *TemplateData) (bytes.Buffer, error)

//go:embed configure.html configure_config.html
var templateFs embed.FS

var ExecuteTemplate = func() TemplateExecutor {
	tmpl := template.Must(template.ParseFS(templateFs, "*.html"))
	return func(data *TemplateData, name string) (bytes.Buffer, error) {
		data.Version = config.Version
		var buf bytes.Buffer
		err := tmpl.ExecuteTemplate(&buf, name, data)
		return buf, err
	}
}()

var GetPage = func() PageGetter {
	return func(data *TemplateData) (bytes.Buffer, error) {
		return ExecuteTemplate(data, "configure.html")
	}
}()
