package stremio

import "encoding/json"

type ResourceName string

const (
	ResourceNameCatalog      ResourceName = "catalog"
	ResourceNameMeta         ResourceName = "meta"
	ResourceNameStream       ResourceName = "stream"
	ResourceNameSubtitles    ResourceName = "subtitles"
	ResourceNameAddonCatalog ResourceName = "addon_catalog"
)

type Resource struct {
	Name       ResourceName  `json:"name"`
	Types      []ContentType `json:"types"`
	IDPrefixes []string      `json:"idPrefixes,omitempty"`
}

type resource Resource

func (r *Resource) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		r.Name = ResourceName(name)
		r.Types = []ContentType{}
		return nil
	}
	rsrc := &resource{}
	err := json.Unmarshal(data, rsrc)
	r.Name = rsrc.Name
	r.Types = rsrc.Types
	r.IDPrefixes = rsrc.IDPrefixes
	return err
}

type CatalogExtra struct {
	Name         string   `json:"name"`
	IsRequired   bool     `json:"isRequired,omitempty"`
	Options      []string `json:"options,omitempty"`
	OptionsLimit int      `json:"optionsLimit,omitempty"`
}

type Catalog struct {
	Type  string         `json:"type"`
	Id    string         `json:"id"`
	Name  string         `json:"name"`
	Extra []CatalogExtra `json:"extra,omitempty"`

	Genres         []string `json:"genres,omitempty"`         //legacy
	ExtraSupported []string `json:"extraSupported,omitempty"` // legacy
	ExtraRequired  []string `json:"extraRequired,omitempty"`  // legacy
}

type BehaviorHints struct {
	Adult                   bool `json:"adult,omitempty"`
	P2P                     bool `json:"p2p,omitempty"`
	Configurable            bool `json:"configurable,omitempty"`
	ConfigurationRequired   bool `json:"configurationRequired,omitempty"`
	NewEpisodeNotifications bool `json:"newEpisodeNotifications,omitempty"` // undocumented
}

type Manifest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`

	Resources  []Resource    `json:"resources"`
	Types      []ContentType `json:"types"`
	IDPrefixes []string      `json:"idPrefixes,omitempty"`

	AddonCatalogs []Catalog `json:"addonCatalogs,omitempty"`
	Catalogs      []Catalog `json:"catalogs"`

	Background    string         `json:"background,omitempty"`
	Logo          string         `json:"logo,omitempty"`
	ContactEmail  string         `json:"contactEmail,omitempty"`
	BehaviorHints *BehaviorHints `json:"behaviorHints,omitempty"`
}

func (m *Manifest) IsValid() bool {
	return m.ID != "" && m.Name != "" && m.Version != ""
}
