package stremio_wrap

import (
	"bytes"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/MunifTanjim/stremthru/internal/kv"
	"github.com/MunifTanjim/stremthru/stremio"
)

type StreamTransformerField string

const (
	StreamTransformerFieldAddon      StreamTransformerField = "addon"
	StreamTransformerFieldBitDepth   StreamTransformerField = "bitdepth"
	StreamTransformerFieldCached     StreamTransformerField = "cached"
	StreamTransformerFieldCodec      StreamTransformerField = "codec"
	StreamTransformerFieldDebrid     StreamTransformerField = "debrid"
	StreamTransformerFieldFileIdx    StreamTransformerField = "fileidx"
	StreamTransformerFieldFilename   StreamTransformerField = "filename"
	StreamTransformerFieldHash       StreamTransformerField = "hash"
	StreamTransformerFieldHDR        StreamTransformerField = "hdr"
	StreamTransformerFieldQuality    StreamTransformerField = "quality"
	StreamTransformerFieldResolution StreamTransformerField = "resolution"
	StreamTransformerFieldSite       StreamTransformerField = "site"
	StreamTransformerFieldSize       StreamTransformerField = "size"
	StreamTransformerFieldTitle      StreamTransformerField = "title"

	StreamTransformerFieldSeason  StreamTransformerField = "season"
	StreamTransformerFieldEpisode StreamTransformerField = "episode"
)

type StreamTransformerExtractorBlob string

type StreamTransformerPattern struct {
	Field string
	Regex *regexp.Regexp
}

type StreamTransformerExtractor []StreamTransformerPattern

func (steb StreamTransformerExtractorBlob) Parse() (StreamTransformerExtractor, error) {
	ste := StreamTransformerExtractor{}
	if steb == "" {
		return ste, nil
	}

	parts := strings.Split(string(steb), "\n")

	field := ""
	lastField := ""
	lastPart := ""
	for _, part := range parts {
		if part == "" && lastPart == "" {
			field = ""
			lastField = ""
			continue
		}
		if field == "" {
			field = part
		} else {
			re, err := regexp.Compile(part)
			if err != nil {
				log.Error("failed to compile regex", "regex", part, "error", err)
				return nil, err
			}
			pattern := StreamTransformerPattern{Regex: re}
			if field != lastField {
				pattern.Field = field
				lastField = field
			}
			ste = append(ste, pattern)
		}
	}

	return ste, nil
}

type StreamTransformerTemplateBlob struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type StreamTransformerTemplate struct {
	Name        *template.Template
	Description *template.Template
}

func (sttb StreamTransformerTemplateBlob) IsEmpty() bool {
	return sttb.Name == "" && sttb.Description == ""
}

func (sttb StreamTransformerTemplateBlob) Parse() (*StreamTransformerTemplate, error) {
	if sttb.IsEmpty() {
		return nil, nil
	}
	stt := &StreamTransformerTemplate{}
	var err error
	stt.Name, err = template.New("name").Parse(sttb.Name)
	if err != nil {
		log.Error("failed to parse name template", "error", err)
		return stt, err
	}
	stt.Description, err = template.New("description").Parse(sttb.Description)
	if err != nil {
		log.Error("failed to parse description template", "error", err)
		return stt, err
	}
	return stt, nil
}

type StreamTransformer struct {
	Extractor StreamTransformerExtractor
	Template  *StreamTransformerTemplate
}

type StreamTransformerResult struct {
	Addon      string
	BitDepth   string
	Codec      string
	Debrid     string
	FileIdx    int
	Filename   string
	HDR        string
	Hash       string
	IsCached   bool
	Quality    string
	Resolution string
	Site       string
	Size       string
	Title      string

	Season  string
	Episode string

	Raw StreamTransformerTemplateBlob
}

func (st StreamTransformer) parse(stream *stremio.Stream) *StreamTransformerResult {
	result := &StreamTransformerResult{FileIdx: -1}
	result.Raw.Name = stream.Name
	result.Raw.Description = stream.Description
	if stream.Description == "" {
		result.Raw.Description = stream.Title
	}

	lastField := ""
	for _, pattern := range st.Extractor {
		field := pattern.Field
		if field == "" {
			field = lastField
		}
		if field == "" {
			continue
		} else {
			lastField = field
		}

		fieldValue := ""
		switch field {
		case "name":
			fieldValue = stream.Name
		case "description":
			fieldValue = stream.Description
			if fieldValue == "" {
				fieldValue = stream.Title
			}
		case "bingeGroup":
			if stream.BehaviorHints != nil {
				fieldValue = stream.BehaviorHints.BingeGroup
			}
		case "url":
			fieldValue = stream.URL
		}
		if fieldValue == "" {
			continue
		}

		for _, match := range pattern.Regex.FindAllStringSubmatch(fieldValue, -1) {
			for i, name := range pattern.Regex.SubexpNames() {
				value := match[i]
				if i != 0 && name != "" && value != "" {
					switch name {
					case "addon":
						result.Addon = value
					case "bitdepth":
						result.BitDepth = value
					case "cached":
						result.IsCached = true
					case "codec":
						result.Codec = value
					case "debrid":
						result.Debrid = value
					case "episode":
						result.Episode = value
					case "fileidx":
						if fileIdx, err := strconv.Atoi(value); err == nil {
							result.FileIdx = fileIdx
						}
					case "filename":
						result.Filename = value
					case "hash":
						result.Hash = value
					case "hdr":
						result.HDR = value
					case "quality":
						result.Quality = value
					case "resolution":
						result.Resolution = value
					case "season":
						result.Season = value
					case "site":
						result.Site = value
					case "size":
						result.Size = value
					case "title":
						result.Title = value
					}
				}
			}
		}
	}

	// normalize
	if result.Filename != "" {
		result.Filename = strings.TrimSpace(result.Filename)
	}
	if result.Resolution != "" {
		result.Resolution = strings.ToUpper(result.Resolution)
	}

	return result
}

type WrappedStream struct {
	*stremio.Stream
	r              *StreamTransformerResult
	noContentProxy bool
}

func (st StreamTransformer) Do(stream *stremio.Stream, tryReconfigure bool) (*WrappedStream, error) {
	s := &WrappedStream{Stream: stream}

	if st.Template == nil {
		return s, nil
	}

	data := st.parse(stream)
	if stream.InfoHash != "" {
		data.Hash = stream.InfoHash
		data.FileIdx = stream.FileIndex
	}
	if stream.BehaviorHints != nil {
		if stream.BehaviorHints.Filename != "" {
			data.Filename = stream.BehaviorHints.Filename
		}
	}

	if tryReconfigure {
		if s.URL != "" && data != nil && data.Hash != "" {
			s.InfoHash = data.Hash
			s.FileIndex = data.FileIdx
			s.URL = ""
			data.Debrid = ""
			data.IsCached = false
			if data.Filename != "" {
				if s.BehaviorHints == nil {
					s.BehaviorHints = &stremio.StreamBehaviorHints{}
				}
				if s.BehaviorHints.Filename == "" {
					s.BehaviorHints.Filename = data.Filename
				}
			}
		}
	}

	s.r = data

	var name bytes.Buffer
	err := st.Template.Name.Execute(&name, data)
	if err != nil {
		return s, err
	}
	stream.Name = name.String()
	var description bytes.Buffer
	err = st.Template.Description.Execute(&description, data)
	if err != nil {
		return s, err
	}
	stream.Description = description.String()
	if stream.Title != "" {
		stream.Title = ""
	}

	return s, nil
}

const BUILTIN_TRANSFORMER_ENTITY_ID_EMOJI = "✨"
const BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX = BUILTIN_TRANSFORMER_ENTITY_ID_EMOJI + " "

var newTransformerExtractorIdMap = map[string]string{
	"Debridio":    BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX + "Debridio",
	"Mediafusion": BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX + "MediaFusion",
	"Torrentio":   BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX + "Torrentio",
}

func getNewTransformerExtractorId(oldId string) string {
	if newId, ok := newTransformerExtractorIdMap[oldId]; ok {
		return newId
	}
	return oldId
}

var builtInExtractors = func() map[string]StreamTransformerExtractorBlob {
	extractors := map[string]StreamTransformerExtractorBlob{}

	extractors[BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX+"Comet"] = StreamTransformerExtractorBlob(strings.TrimSpace(`
name
(?i)^\[(?:TORRENT🧲|(?<debrid>\w+)(?<cached>⚡)?)\] (?<addon>.+) (?:unknown|(?<resolution>\d[^kp]*[kp]))

description
^(?<title>.+)\n(?:💿 .+\n)?(?:👤 \d+ )?💾 (?:(?<size>[\d.]+ [^ ]+)|.+?) 🔎 (?<site>.+)(?:\n.+)?
(?i)💿 (?:.+\|)?(?<quality>cam|scr|dvd|vhs|r5|(?:[\w ]+(?:rip|ray|mux|tv))|(?:(?:tele|web)[\w-][\w]+))
(?i)💿 (?:.+\|)?(?<codec>hevc|avc|mpeg|xvid|av1|x264|x265|h264|h265)

url
\/playback\/(?<hash>[a-f0-9]{40})\/(?:n|(?<fileidx>\d+))\/[^/]+\/(?:n|(?<season>\d+))\/(?:n|(?<episode>\d+))\/(?<filename>.+)
`))

	extractors[BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX+"Debridio"] = StreamTransformerExtractorBlob(strings.TrimSpace(`
name
(?i)^(?:\[(?<debrid>\w+?)(?<cached>\+?)\] \n)?(?<addon>\w+) (?:Other|(?<resolution>\d[^kp]*[kp]))

description
^(?<title>.+?) ?\n(?:(?<filename>.+?) ?\n)?⚡? 📺 (?<resolution>[^ ]+) 💾 (?:Unknown|(?<size>[\d.]+ [^ ]+)|.+?) (?:👤 (?:Unknown|\d+))? ⚙️ (?<site>[^ ]+)

url
\/(?<hash>[a-f0-9]{40})(?:\/(?<season>\d+)\/(?<episode>\d+))?
`))

	extractors[BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX+"MediaFusion"] = StreamTransformerExtractorBlob(strings.TrimSpace(`
name
(?i)^(?<addon>\w+(?: \| [^ ]+)?) (?:P2P|(?<debrid>[A-Z]{2})) (?:N\/A|(?<resolution>[^kp]+[kp])) (?<cached>⚡️)?

description
(?i)(?:📂 (?<title>.+)\n)?💾 (?<size>.+?)(?: \/ 💾 .+?)(?: 👤 \d+)?\n(?:.+\n)?🔗 (?<site>.+?)(?: 🧑‍💻 |$)

bingeGroup
(?i)-(?:🎨 (?<hdr>[^ ]+) )?📺 (?<quality>cam|scr|dvd|vhs|r5|(?:.+(?:rip|ray|mux|tv))|(?:(?:tele|web)[\w-]*?))(?: ?🎞️ (?<codec>[^- ]+))?(?: ?🎵 .+)?-(?:N\/A|(?:\d+[kp]))

url
\/stream\/(?<hash>[a-f0-9]{40})\/
`))

	extractors[BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX+"Peerflix"] = StreamTransformerExtractorBlob(strings.TrimSpace(`
name
(?i)^(?:\[(?<debrid>\w+?)(?:(?<cached>\+?)|\s[^\]]+)\] )?(?<addon>\w+) \S+ (?<resolution>\d+[kp])?

description
^(?<title>.+)\n(?<filename>.+\n)?.+👤 \d+ (?:💾 (?<size>[\d.]+ \w[bB]) )?🌐 (?<site>\w+)$
`))

	extractors[BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX+"Torrentio"] = StreamTransformerExtractorBlob(strings.TrimSpace(`
name
(?i)^(?:\[(?<debrid>\w+?)(?<cached>\+?)\] )?(?<addon>\w+)\n?(?<resolution>[^kp]+[kp])?(?: 3D(?: SBS))?(?: (?<hdr>.+))?

description
^(?<title>.+)\n(?:(?<filename>[^👤].+)\n)?👤.+ 💾 (?<size>.+) ⚙️ (?<site>\w+)(?:\n(?<lang>.+))?$
(?i)(?<quality>cam|scr|dvd|vhs|r5|(?:(?:bd|blu|hd|sd)(?:rip|ray|mux|tv))|(?:(?:tele|web)[\w-][\w]+))

bingeGroup
(?i)(?<codec>hevc|avc|mpeg|xvid|av1|x264|x265|h264|h265)
(?i)(?<bitdepth>\d+bit)
(?i)(?<quality>cam|scr|dvd|vhs|r5|(?:\w+(?:rip|ray|mux|tv))|(?:(?:tele|web)[\w-]+))

url
(?i)\/(?<hash>[a-f0-9]{40})\/[^/]+\/(?<fileidx>\d+)\/
`))

	extractors[BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX+"Orion"] = StreamTransformerExtractorBlob(strings.TrimSpace(`
name
(?:🪐 (?<addon>\w+) 📺 (?<resolution>\w+))|(?:(?<cached>🚀) (?<addon>\w+)\n\[(?<debrid>[^\]]+)\])

description
(?<title>.+)\n(?:📺(?<resolution>.+?) )?💾(?<size>[0-9.]+ [^ ]+) (?:👤\d+ )?🎥(?<codec>\w+) 🔊.+\n👂.+ ☁️(?<site>.+)
`))

	return extractors
}()

var extractorStore = func() kv.KVStore[StreamTransformerExtractorBlob] {
	return kv.NewKVStore[StreamTransformerExtractorBlob](&kv.KVStoreConfig{
		Type: "st:wrap:transformer:extractor",
		GetKey: func(key string) string {
			return key
		},
	})
}()

func getExtractor(extractorId string) (StreamTransformerExtractorBlob, error) {
	if strings.HasPrefix(extractorId, BUILTIN_TRANSFORMER_ENTITY_ID_EMOJI) {
		if extractor, ok := builtInExtractors[extractorId]; ok {
			return extractor, nil
		}
		return "", errors.New("built-in extractor not found")
	}

	var extractor StreamTransformerExtractorBlob
	if err := extractorStore.Get(extractorId, &extractor); err != nil {
		return "", err
	}
	return extractor, nil
}

func getExtractorIds() ([]string, error) {
	extractors, err := extractorStore.List()
	if err != nil {
		return nil, err
	}
	builtInExtractorsCount := len(builtInExtractors)
	extractorIds := make([]string, builtInExtractorsCount+len(extractors))
	idx := 0
	for id := range builtInExtractors {
		extractorIds[idx] = id
		idx++
	}
	for _, extractor := range extractors {
		extractorIds[idx] = extractor.Key
		idx++
	}
	return extractorIds, nil

}

var builtInTemplates = func() map[string]StreamTransformerTemplateBlob {
	templates := map[string]StreamTransformerTemplateBlob{}

	templates[BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX+"Default"] = StreamTransformerTemplateBlob{
		Name: strings.TrimSpace(`
{{if ne .Debrid ""}}[{{if .IsCached}}⚡️{{end}}{{.Debrid}}]
{{end}}{{.Addon}}
{{.Resolution}}
`),
		Description: strings.TrimSpace(`
{{if ne .Quality ""}}🎥 {{.Quality}} {{end}}{{if ne .Codec ""}}🎞️ {{.Codec}}{{end}}
{{if ne .Size ""}}📦 {{.Size}} {{end}}{{if ne .HDR ""}}📺 {{.HDR}} {{end}}{{if ne .Site ""}}🔗 {{.Site}}{{end}}{{if ne .Filename ""}}
📄 {{.Filename}}{{else if ne .Title ""}}
📁 {{.Title}}
{{end}}
`),
	}

	templates[BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX+"Raw"] = StreamTransformerTemplateBlob{
		Name:        `{{.Raw.Name}}`,
		Description: `{{.Raw.Description}}`,
	}

	return templates
}()

var templateStore = func() kv.KVStore[StreamTransformerTemplateBlob] {
	return kv.NewKVStore[StreamTransformerTemplateBlob](&kv.KVStoreConfig{
		Type: "st:wrap:transformer:template",
		GetKey: func(key string) string {
			return key
		},
	})
}()

func getTemplate(templateId string) (StreamTransformerTemplateBlob, error) {
	if strings.HasPrefix(templateId, BUILTIN_TRANSFORMER_ENTITY_ID_EMOJI) {
		if template, ok := builtInTemplates[templateId]; ok {
			return template, nil
		}
		return StreamTransformerTemplateBlob{}, errors.New("built-in template not found")
	}

	var template StreamTransformerTemplateBlob
	if err := templateStore.Get(templateId, &template); err != nil {
		return StreamTransformerTemplateBlob{}, err
	}
	return template, nil
}

func getTemplateIds() ([]string, error) {
	templates, err := templateStore.List()
	if err != nil {
		return nil, err
	}
	builtInTemplatesCount := len(builtInTemplates)
	templateIds := make([]string, builtInTemplatesCount+len(templates))
	idx := 0
	for id := range builtInTemplates {
		templateIds[idx] = id
		idx++
	}
	for _, template := range templates {
		templateIds[idx] = template.Key
		idx++
	}
	return templateIds, nil
}

func seedDefaultTransformerEntities() {
	if config.IsPublicInstance {
		for oldId := range newTransformerExtractorIdMap {
			if err := extractorStore.Del(oldId); err != nil {
				log.Warn("Failed to cleanup seed extractor: " + oldId)
			}
		}
	}
	for id := range builtInExtractors {
		if err := extractorStore.Del(id); err != nil {
			log.Warn("Failed to cleanup seed extractor: " + id)
		}
	}

	for key := range builtInTemplates {
		if err := templateStore.Del(key); err != nil {
			log.Warn("Failed to cleanup seed template: " + key)
		}
		if config.IsPublicInstance {
			key = strings.TrimPrefix(key, BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX)
			if err := templateStore.Del(key); err != nil {
				log.Warn("Failed to cleanup seed template: " + key)
			}
		}
	}
}
