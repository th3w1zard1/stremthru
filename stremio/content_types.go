package stremio

type ContentType string

const (
	ContentTypeMovie   ContentType = "movie"
	ContentTypeSeries  ContentType = "series"
	ContentTypeChannel ContentType = "channel"
	ContentTypeTV      ContentType = "tv"
)
