package buddy

import (
	"time"

	"github.com/MunifTanjim/stremthru/core"
	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/MunifTanjim/stremthru/internal/logger"
	"github.com/MunifTanjim/stremthru/internal/peer"
	ti "github.com/MunifTanjim/stremthru/internal/torrent_info"
	tss "github.com/MunifTanjim/stremthru/internal/torrent_stream/torrent_stream_syncinfo"
)

var PullPeer, pullLocalOnly = func() (*peer.APIClient, bool) {
	baseUrl := config.PullPeerURL
	if baseUrl == "" {
		baseUrl = config.PeerURL
	}
	localOnly := baseUrl == config.PullPeerURL
	if baseUrl == "" {
		return nil, localOnly
	}
	return peer.NewAPIClient(&peer.APIClientConfig{
		BaseURL: baseUrl,
	}), localOnly
}()

var pullPeerLog = logger.Scoped("peer:pull")

func PullTorrentsByStremId(sid string, originInstanceId string) {
	if PullPeer == nil || !tss.ShouldPull(sid) {
		return
	}

	start := time.Now()
	res, err := PullPeer.ListTorrents(&peer.ListTorrentsByStremIdParams{
		SId:              sid,
		LocalOnly:        pullLocalOnly,
		OriginInstanceId: originInstanceId,
	})
	duration := time.Since(start)

	if err != nil {
		pullPeerLog.Error("failed to pull torrents", "error", core.PackError(err), "duration", duration, "sid", sid)
		return
	}

	count := len(res.Data.Items)
	pullPeerLog.Info("pulled torrents", "duration", duration, "sid", sid, "count", count)

	items := make([]ti.TorrentInfoInsertData, count)
	for i := range res.Data.Items {
		data := &res.Data.Items[i]
		items[i] = ti.TorrentInfoInsertData{
			Hash:         data.Hash,
			TorrentTitle: data.TorrentTitle,
			Size:         data.Size,
			Source:       ti.TorrentInfoSource(data.Source),
			Category:     ti.TorrentInfoCategory(data.Category),
			Files:        data.Files,
		}
	}
	ti.Upsert(items, "", false)
	go tss.MarkPulled(sid)
}

func ListTorrentsByStremId(sid string, localOnly bool, originInstanceId string) (*ti.ListTorrentsData, error) {
	if originInstanceId == config.InstanceId && !pullLocalOnly {
		pullPeerLog.Info("loop detected for list torrents, self-correcting...")
		pullLocalOnly = true
	}

	if !localOnly {
		PullTorrentsByStremId(sid, originInstanceId)
	}

	data, err := ti.ListByStremId(sid)
	if err != nil {
		return nil, err
	}
	return data, nil
}
