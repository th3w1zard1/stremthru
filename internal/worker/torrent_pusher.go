package worker

import (
	"errors"
	"time"

	"github.com/MunifTanjim/stremthru/core"
	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/MunifTanjim/stremthru/internal/logger"
	"github.com/MunifTanjim/stremthru/internal/peer"
	"github.com/MunifTanjim/stremthru/internal/torrent_info"
	tss "github.com/MunifTanjim/stremthru/internal/torrent_stream/torrent_stream_syncinfo"
	"github.com/madflojo/tasks"
)

var TorrentPusherQueue = IdQueue{
	debounceTime: 5 * time.Minute,
}

var Peer = peer.NewAPIClient(&peer.APIClientConfig{
	BaseURL: config.PeerURL,
	APIKey:  config.PeerAuthToken,
})

func InitPushTorrentsWorker() *tasks.Scheduler {
	if !config.HasPeer || config.PeerAuthToken == "" {
		return nil
	}

	log := logger.Scoped("worker/torrent_pusher")

	scheduler := tasks.New()

	id, err := scheduler.Add(&tasks.Task{
		Interval:          time.Duration(10 * time.Minute),
		RunSingleInstance: true,
		TaskFunc: func() (err error) {
			defer func() {
				if e := recover(); e != nil {
					if pe, ok := e.(error); ok {
						err = pe
					} else {
						err = errors.New("something went wrong")
					}
				}
			}()

			TorrentPusherQueue.m.Range(func(k, v any) bool {
				sid, sidOk := k.(string)
				t, tOk := v.(time.Time)
				if sidOk && tOk && t.Before(time.Now()) {
					if tss.ShouldPush(sid) {
						if data, err := torrent_info.ListByStremId(sid); err == nil {
							params := &peer.PushTorrentsParams{
								Items: data.Items,
							}
							start := time.Now()
							if _, err := Peer.PushTorrents(params); err != nil {
								log.Error("failed to push torrents", "error", core.PackError(err), "duration", time.Since(start), "count", data.TotalItems)
							} else {
								log.Info("pushed torrents", "duration", time.Since(start), "count", data.TotalItems)
								tss.MarkPushed(sid)
							}
						} else {
							log.Error("failed to list torrents", "error", core.PackError(err), "sid", sid)
						}

					}

					TorrentPusherQueue.delete(sid)
				}
				return true
			})

			return nil
		},
		ErrFunc: func(err error) {
			log.Error("Worker Failure", "error", err)
		},
	})

	if err != nil {
		panic(err)
	}

	log.Info("Started Worker", "id", id)

	return scheduler
}
