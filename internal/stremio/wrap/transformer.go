package stremio_wrap

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/MunifTanjim/stremthru/internal/kv"
	"github.com/MunifTanjim/stremthru/internal/shared"
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
)

var extractorStore = func() kv.KVStore[StreamTransformerExtractorBlob] {
	return kv.NewKVStore[StreamTransformerExtractorBlob](&kv.KVStoreConfig{
		Type: "st:wrap:transformer:extractor",
		GetKey: func(key string) string {
			return key
		},
	})
}()

type StreamTransformerExtractorBlob string

type StreamTransformerPattern struct {
	Field string
	Regex *regexp.Regexp
}

type StreamTransformerExtractor []StreamTransformerPattern

func (steb StreamTransformerExtractorBlob) Parse() (StreamTransformerExtractor, error) {
	ste := StreamTransformerExtractor{}
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
				log.Printf("[stremio/wrap] failed to compile regex %s: %v", part, err)
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

var templateStore = func() kv.KVStore[StreamTransformerTemplateBlob] {
	return kv.NewKVStore[StreamTransformerTemplateBlob](&kv.KVStoreConfig{
		Type: "st:wrap:transformer:template",
		GetKey: func(key string) string {
			return key
		},
	})
}()

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
	if sttb.Name == "" && sttb.Description == "" {
		return nil, nil
	}
	stt := &StreamTransformerTemplate{}
	var err error
	stt.Name, err = template.New("name").Parse(sttb.Name)
	if err != nil {
		log.Printf("[stremio/wrap] failed to parse name template: %v", err)
		return stt, err
	}
	stt.Description, err = template.New("description").Parse(sttb.Description)
	if err != nil {
		log.Printf("[stremio/wrap] failed to parse description template: %v", err)
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
}

func (str StreamTransformerResult) GetResolutionRank() int {
	if strings.HasSuffix(str.Resolution, "P") {
		if resolution, err := strconv.Atoi(str.Resolution[:len(str.Resolution)-1]); err == nil {
			return resolution
		}
	}
	if strings.HasSuffix(str.Resolution, "K") {
		if resolution, err := strconv.Atoi(str.Resolution[:len(str.Resolution)-1]); err == nil {
			return resolution * 1000
		}
	}
	return 0
}

func (str StreamTransformerResult) GetQualityRank() int {
	quality := strings.ToLower(str.Quality)

	if strings.Contains(quality, "mux") {
		return 100
	}

	if strings.Contains(quality, "bluray") {
		return 99
	}
	if strings.Contains(quality, "br") {
		return 97
	}
	if strings.Contains(quality, "bd") {
		return 95
	}
	if strings.Contains(quality, "uhd") {
		return 93
	}

	if strings.Contains(quality, "web") {
		if strings.Contains(quality, "dl") {
			return 89
		}
		if strings.Contains(quality, "rip") {
			return 85
		}
		return 80
	}

	if strings.Contains(quality, "hd") {
		if strings.Contains(quality, "rip") {
			return 79
		}
		return 75
	}

	if strings.Contains(quality, "dvd") {
		return 60
	}

	if strings.Contains(quality, "sat") {
		return 59
	}
	if strings.Contains(quality, "tv") {
		return 57
	}
	if strings.Contains(quality, "ppv") {
		return 55
	}

	if strings.Contains(quality, "cam") {
		return 40
	}
	if strings.Contains(quality, "tele") {
		return 30
	}
	if strings.Contains(quality, "scr") {
		return 20
	}
	return 0
}

func (str StreamTransformerResult) GetSizeRank() int64 {
	return shared.ToBytes(str.Size)
}

func (st StreamTransformer) parse(stream *stremio.Stream) *StreamTransformerResult {
	result := &StreamTransformerResult{}
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
				if i != 0 && name != "" {
					switch name {
					case "addon":
						result.Addon = match[i]
					case "bitdepth":
						result.BitDepth = match[i]
					case "cached":
						result.IsCached = match[i] != ""
					case "codec":
						result.Codec = match[i]
					case "debrid":
						result.Debrid = match[i]
					case "fileidx":
						if fileIdx, err := strconv.Atoi(match[i]); err == nil {
							result.FileIdx = fileIdx
						}
					case "filename":
						result.Filename = match[i]
					case "hash":
						result.Hash = match[i]
					case "hdr":
						result.HDR = match[i]
					case "quality":
						result.Quality = match[i]
					case "resolution":
						result.Resolution = match[i]
					case "site":
						result.Site = match[i]
					case "size":
						result.Size = match[i]
					case "title":
						result.Title = match[i]
					}
				}
			}
		}
	}
	if result.Resolution != "" {
		result.Resolution = strings.ToUpper(result.Resolution)
	}
	return result
}

type WrappedStream struct {
	*stremio.Stream
	r *StreamTransformerResult
}

func (st StreamTransformer) Do(stream *stremio.Stream) (*WrappedStream, error) {
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

type StreamSortConfig struct {
	Field string
	Desc  bool
}

func parseSortConfig(config string) []StreamSortConfig {
	sortConfigs := []StreamSortConfig{}
	for _, part := range strings.Split(config, ",") {
		part = strings.TrimSpace(part)
		desc := strings.HasPrefix(part, "-")
		field := strings.TrimPrefix(part, "-")
		if field == "resolution" || field == "quality" || field == "size" {
			sortConfigs = append(sortConfigs, StreamSortConfig{Field: field, Desc: desc})
		}
	}
	return sortConfigs
}

func getFieldValue(ws WrappedStream, field string) interface{} {
	switch field {
	case "resolution":
		return ws.r.GetResolutionRank()
	case "quality":
		return ws.r.GetQualityRank()
	case "size":
		return ws.r.GetSizeRank()
	}
	panic("Unsupported field type for sorting")
}

type StreamSorter struct {
	items  []WrappedStream
	config []StreamSortConfig
}

func (ss StreamSorter) Len() int {
	return len(ss.items)
}
func (ss StreamSorter) Swap(i, j int) {
	ss.items[i], ss.items[j] = ss.items[j], ss.items[i]
}

func (ss StreamSorter) Less(a, b int) bool {
	for _, config := range ss.config {
		va := getFieldValue(ss.items[a], config.Field)
		vb := getFieldValue(ss.items[b], config.Field)

		if va == vb {
			continue
		}

		switch vaTyped := va.(type) {
		case int:
			vbTyped := vb.(int)
			if config.Desc {
				return vaTyped > vbTyped
			}
			return vaTyped < vbTyped
		case int64:
			vbTyped := vb.(int64)
			if config.Desc {
				return vaTyped > vbTyped
			}
			return vaTyped < vbTyped
		default:
			panic(fmt.Sprintf("Unsupported field type for sorting: %T", va))
		}
	}
	return false
}

func SortWrappedStreams(items []WrappedStream, config string) {
	if config == "" {
		config = "-resolution,-quality,-size"
	}

	sortConfigs := parseSortConfig(config)
	if len(sortConfigs) == 0 {
		return
	}
	sorter := StreamSorter{items: items, config: sortConfigs}
	sort.Sort(sorter)
}
