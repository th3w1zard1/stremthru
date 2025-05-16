package stremio_shared

import (
	"regexp"
	"strings"

	"github.com/MunifTanjim/stremthru/store"
)

func MatchFileByIdx(files []store.MagnetFile, idx int) *store.MagnetFile {
	if idx == -1 {
		return nil
	}
	for i := range files {
		f := &files[i]
		if f.Idx == idx {
			return f
		}
	}
	return nil
}

func MatchFileByLargestSize(files []store.MagnetFile) (file *store.MagnetFile) {
	for i := range files {
		f := &files[i]
		if file == nil || file.Size < f.Size {
			file = f
		}
	}
	return file
}

func MatchFileByName(files []store.MagnetFile, name string) *store.MagnetFile {
	if name == "" {
		return nil
	}
	for i := range files {
		f := &files[i]
		if f.Name == name {
			return f
		}
	}
	return nil
}

func MatchFileByPattern(files []store.MagnetFile, pattern *regexp.Regexp) *store.MagnetFile {
	if pattern == nil {
		return nil
	}
	for i := range files {
		f := &files[i]
		if pattern.MatchString(f.Name) {
			return f
		}
	}
	return nil
}

func MatchFileByStremId(files []store.MagnetFile, sid string) *store.MagnetFile {
	if parts := strings.SplitN(sid, ":", 3); len(parts) == 3 {
		if pat, err := regexp.Compile("0?" + parts[1] + `\D{1,3}` + "0?" + parts[2]); err == nil {
			for i := range files {
				f := &files[i]
				if pat.MatchString(f.Name) {
					return f
				}
			}
		}
	}
	return nil
}
