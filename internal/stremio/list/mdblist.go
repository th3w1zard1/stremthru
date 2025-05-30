package stremio_list

import (
	"github.com/MunifTanjim/stremthru/internal/mdblist"
)

var mdblistClient = mdblist.NewAPIClient(&mdblist.APIClientConfig{})
