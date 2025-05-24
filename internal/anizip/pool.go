package anizip

import (
	"sync"

	"github.com/alitto/pond/v2"
)

var GetMappingsPool = sync.OnceValue(func() pond.ResultPool[*GetMappingsData] {
	return pond.NewResultPool[*GetMappingsData](10)
})
