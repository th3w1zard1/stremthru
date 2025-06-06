package worker_queue

import (
	"time"

	"github.com/MunifTanjim/stremthru/internal/config"
)

type AnimeIdMapperQueueItem struct {
	Service string
	Id      string
}

var AnimeIdMapperQueue = WorkerQueue[AnimeIdMapperQueueItem]{
	debounceTime: 1 * time.Minute,
	getKey: func(item AnimeIdMapperQueueItem) string {
		return item.Service + ":" + item.Id
	},
	getGroupKey: func(item AnimeIdMapperQueueItem) string {
		return item.Service
	},
	transform: func(item *AnimeIdMapperQueueItem) *AnimeIdMapperQueueItem {
		return item
	},
	Disabled: !config.Feature.IsEnabled("anime"),
}
