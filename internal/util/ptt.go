package util

import (
	"github.com/MunifTanjim/go-ptt"
)

func ParseTorrentTitle(title string) (*ptt.Result, error) {
	r := ptt.Parse(title)
	if err := r.Error(); err != nil {
		return r, err
	}
	return r.Normalize(), nil
}
