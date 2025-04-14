package torrent_info

import (
	"testing"

	"github.com/MunifTanjim/stremthru/stremio"
	"github.com/stretchr/testify/assert"
)

func TestExtractorTorrentioTorrent(t *testing.T) {
	for _, tc := range []struct {
		name     string
		hostname string
		sid      string
		stream   stremio.Stream
		data     TorrentInfoInsertData
	}{
		{
			"single",
			"torrentio.strem.fun", "tt1431045", stremio.Stream{
				Name:      "Torrentio\n4k DV",
				Title:     "Deadpool.2016.UHD.BluRay.2160p.TrueHD.Atmos.7.1.DV.HEVC.HYBRiD.REMUX-FraMeSToR\n👤 47 💾 40.33 GB ⚙️ TorrentGalaxy",
				InfoHash:  "e4f5d7a2f3dd6b7b1826bd77e316b6b5ba31eb72",
				FileIndex: 0,
				BehaviorHints: &stremio.StreamBehaviorHints{
					BingeGroup: "torrentio|4k|BluRay REMUX|hevc|DV",
					Filename:   "Deadpool.2016.UHD.BluRay.2160p.TrueHD.Atmos.7.1.DV.HEVC.HYBRiD.REMUX-FraMeSToR.mkv",
				},
			}, TorrentInfoInsertData{
				Hash:         "e4f5d7a2f3dd6b7b1826bd77e316b6b5ba31eb72",
				TorrentTitle: "Deadpool.2016.UHD.BluRay.2160p.TrueHD.Atmos.7.1.DV.HEVC.HYBRiD.REMUX-FraMeSToR",
				Size:         -1,
				Source:       "tio",
				Files: []TorrentInfoInsertDataFile{
					{
						Name:   "Deadpool.2016.UHD.BluRay.2160p.TrueHD.Atmos.7.1.DV.HEVC.HYBRiD.REMUX-FraMeSToR.mkv",
						Idx:    0,
						Size:   43304007761,
						SId:    "tt1431045",
						Source: "",
					},
				},
			},
		},
		{
			"multi w/ behaviorHints.filename",
			"torrentio.strem.fun", "tt1431045", stremio.Stream{
				Name:      "Torrentio\n1080p",
				Title:     "X-Men Complete 13 Movie Collection Sci-Fi 2000 - 2020 Eng Rus Multi-Subs 1080p [H264-mp4]\nX-Men Complete 13 Movie Collection 2000-2020/08 Deadpool - Action 2016 Eng Rus Multi-Subs 1080p [H264-mp4].mp4\n👤 20 💾 3.65 GB ⚙️ TorrentGalaxy\nMulti Subs / 🇬🇧 / 🇷🇺 / 🇫🇮",
				InfoHash:  "a6a80257d62e53e55c877a7067ea5055129b462c",
				FileIndex: 89,
				BehaviorHints: &stremio.StreamBehaviorHints{
					BingeGroup: "torrentio|1080p|h264",
					Filename:   "08 Deadpool - Action 2016 Eng Rus Multi-Subs 1080p [H264-mp4].mp4",
				},
			}, TorrentInfoInsertData{
				Hash:         "a6a80257d62e53e55c877a7067ea5055129b462c",
				TorrentTitle: "X-Men Complete 13 Movie Collection Sci-Fi 2000 - 2020 Eng Rus Multi-Subs 1080p [H264-mp4]",
				Size:         -1,
				Source:       "tio",
				Files: []TorrentInfoInsertDataFile{
					{
						Name:   "08 Deadpool - Action 2016 Eng Rus Multi-Subs 1080p [H264-mp4].mp4",
						Idx:    89,
						Size:   3919157657,
						SId:    "tt1431045",
						Source: "",
					},
				},
			},
		},
		{
			"multi w/o behaviorHints.filename",
			"torrentio.strem.fun", "tt1431045", stremio.Stream{
				Name:      "Torrentio\n1080p",
				Title:     "X-Men Complete 13 Movie Collection Sci-Fi 2000 - 2020 Eng Rus Multi-Subs 1080p [H264-mp4]\nX-Men Complete 13 Movie Collection 2000-2020/08 Deadpool - Action 2016 Eng Rus Multi-Subs 1080p [H264-mp4].mp4\n👤 20 💾 3.65 GB ⚙️ TorrentGalaxy\nMulti Subs / 🇬🇧 / 🇷🇺 / 🇫🇮",
				InfoHash:  "a6a80257d62e53e55c877a7067ea5055129b462c",
				FileIndex: 89,
				BehaviorHints: &stremio.StreamBehaviorHints{
					BingeGroup: "torrentio|1080p|h264",
				},
			}, TorrentInfoInsertData{
				Hash:         "a6a80257d62e53e55c877a7067ea5055129b462c",
				TorrentTitle: "X-Men Complete 13 Movie Collection Sci-Fi 2000 - 2020 Eng Rus Multi-Subs 1080p [H264-mp4]",
				Size:         -1,
				Source:       "tio",
				Files: []TorrentInfoInsertDataFile{
					{
						Name:   "08 Deadpool - Action 2016 Eng Rus Multi-Subs 1080p [H264-mp4].mp4",
						Idx:    89,
						Size:   3919157657,
						SId:    "tt1431045",
						Source: "",
					},
				},
			},
		},
		{
			"missing filename",
			"torrentio.strem.fun", "tt1431045", stremio.Stream{
				Name:     "Torrentio\n720p",
				Title:    "Deadpool (2016) 720p BluRay x264 [Dual Audio] [Hindi (Line Audio) - English] ESubs - Downloadhub\n👤 3 💾 934.8 MB ⚙️ 1337x\nDual Audio / 🇬🇧 / 🇮🇳",
				InfoHash: "f5d0ab292f5a244a4b38efac9ae1f8d311179588",
				BehaviorHints: &stremio.StreamBehaviorHints{
					BingeGroup: "torrentio|720p|BluRay|x264",
				},
			}, TorrentInfoInsertData{
				Hash:         "f5d0ab292f5a244a4b38efac9ae1f8d311179588",
				TorrentTitle: "Deadpool (2016) 720p BluRay x264 [Dual Audio] [Hindi (Line Audio) - English] ESubs - Downloadhub",
				Size:         -1,
				Source:       "tio",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := ExtractCreateDataFromStream(tc.hostname, tc.sid, &tc.stream)
			assert.Equal(t, &tc.data, data)
		})
	}
}

func TestExtractorTorrentioDebrid(t *testing.T) {
	for _, tc := range []struct {
		name     string
		hostname string
		sid      string
		stream   stremio.Stream
		data     TorrentInfoInsertData
	}{
		{
			"single",
			"torrentio.strem.fun", "tt1431045", stremio.Stream{
				Name:  "[RD+] Torrentio\n4k DV",
				Title: "Deadpool.2016.UHD.BluRay.2160p.TrueHD.Atmos.7.1.DV.HEVC.HYBRiD.REMUX-FraMeSToR\n👤 47 💾 40.33 GB ⚙️ TorrentGalaxy",
				URL:   "https://torrentio.strem.fun/realdebrid/xxxxxxxxxxxxxxxx/e4f5d7a2f3dd6b7b1826bd77e316b6b5ba31eb72/null/0/Deadpool.2016.UHD.BluRay.2160p.TrueHD.Atmos.7.1.DV.HEVC.HYBRiD.REMUX-FraMeSToR.mkv",
				BehaviorHints: &stremio.StreamBehaviorHints{
					BingeGroup: "torrentio|4k|BluRay REMUX|hevc|DV",
					Filename:   "Deadpool.2016.UHD.BluRay.2160p.TrueHD.Atmos.7.1.DV.HEVC.HYBRiD.REMUX-FraMeSToR.mkv",
				},
			}, TorrentInfoInsertData{
				Hash:         "e4f5d7a2f3dd6b7b1826bd77e316b6b5ba31eb72",
				TorrentTitle: "Deadpool.2016.UHD.BluRay.2160p.TrueHD.Atmos.7.1.DV.HEVC.HYBRiD.REMUX-FraMeSToR",
				Size:         -1,
				Source:       "tio",
				Files: []TorrentInfoInsertDataFile{
					{
						Name:   "Deadpool.2016.UHD.BluRay.2160p.TrueHD.Atmos.7.1.DV.HEVC.HYBRiD.REMUX-FraMeSToR.mkv",
						Idx:    0,
						Size:   43304007761,
						SId:    "tt1431045",
						Source: "",
					},
				},
			},
		},
		{
			"single - untrusted fileidx",
			"torrentio.strem.fun", "tt1431045", stremio.Stream{
				Name:  "[AD download] Torrentio\n4k DV",
				Title: "Deadpool.2016.UHD.BluRay.2160p.TrueHD.Atmos.7.1.DV.HEVC.HYBRiD.REMUX-FraMeSToR\n👤 47 💾 40.33 GB ⚙️ TorrentGalaxy",
				URL:   "https://torrentio.strem.fun/realdebrid/xxxxxxxxxxxxxxxx/e4f5d7a2f3dd6b7b1826bd77e316b6b5ba31eb72/null/0/Deadpool.2016.UHD.BluRay.2160p.TrueHD.Atmos.7.1.DV.HEVC.HYBRiD.REMUX-FraMeSToR.mkv",
				BehaviorHints: &stremio.StreamBehaviorHints{
					BingeGroup: "torrentio|4k|BluRay REMUX|hevc|DV",
					Filename:   "Deadpool.2016.UHD.BluRay.2160p.TrueHD.Atmos.7.1.DV.HEVC.HYBRiD.REMUX-FraMeSToR.mkv",
				},
			}, TorrentInfoInsertData{
				Hash:         "e4f5d7a2f3dd6b7b1826bd77e316b6b5ba31eb72",
				TorrentTitle: "Deadpool.2016.UHD.BluRay.2160p.TrueHD.Atmos.7.1.DV.HEVC.HYBRiD.REMUX-FraMeSToR",
				Size:         -1,
				Source:       "tio",
				Files: []TorrentInfoInsertDataFile{
					{
						Name:   "Deadpool.2016.UHD.BluRay.2160p.TrueHD.Atmos.7.1.DV.HEVC.HYBRiD.REMUX-FraMeSToR.mkv",
						Idx:    -1,
						Size:   43304007761,
						SId:    "tt1431045",
						Source: "",
					},
				},
			},
		},
		{
			"multi",
			"torrentio.strem.fun", "tt1431045", stremio.Stream{
				Name:  "[RD+] Torrentio\n1080p",
				Title: "X-Men Complete 13 Movie Collection Sci-Fi 2000 - 2020 Eng Rus Multi-Subs 1080p [H264-mp4]\nX-Men Complete 13 Movie Collection 2000-2020/08 Deadpool - Action 2016 Eng Rus Multi-Subs 1080p [H264-mp4].mp4\n👤 20 💾 3.65 GB ⚙️ TorrentGalaxy\nMulti Subs / 🇬🇧 / 🇷🇺 / 🇫🇮",
				URL:   "https://torrentio.strem.fun/realdebrid/xxxxxxxxxxxxxxxx/a6a80257d62e53e55c877a7067ea5055129b462c/null/89/08%20Deadpool%20-%20Action%202016%20Eng%20Rus%20Multi-Subs%201080p%20%5BH264-mp4%5D.mp4",
				BehaviorHints: &stremio.StreamBehaviorHints{
					BingeGroup: "torrentio|1080p|h264",
					Filename:   "08 Deadpool - Action 2016 Eng Rus Multi-Subs 1080p [H264-mp4].mp4",
				},
			}, TorrentInfoInsertData{
				Hash:         "a6a80257d62e53e55c877a7067ea5055129b462c",
				TorrentTitle: "X-Men Complete 13 Movie Collection Sci-Fi 2000 - 2020 Eng Rus Multi-Subs 1080p [H264-mp4]",
				Size:         -1,
				Source:       "tio",
				Files: []TorrentInfoInsertDataFile{
					{
						Name:   "08 Deadpool - Action 2016 Eng Rus Multi-Subs 1080p [H264-mp4].mp4",
						Idx:    89,
						Size:   3919157657,
						SId:    "tt1431045",
						Source: "",
					},
				},
			},
		},
		{
			"missing filename",
			"torrentio.strem.fun", "tt1431045", stremio.Stream{
				Name:  "[RD download] Torrentio\n720p",
				Title: "Deadpool (2016) 720p BluRay x264 [Dual Audio] [Hindi (Line Audio) - English] ESubs - Downloadhub\n👤 3 💾 934.8 MB ⚙️ 1337x\nDual Audio / 🇬🇧 / 🇮🇳",
				URL:   "https://torrentio.strem.fun/realdebrid/xxxxxxxxxxxxxxxx/f5d0ab292f5a244a4b38efac9ae1f8d311179588/null/undefined/Deadpool%20(2016)%20720p%20BluRay%20x264%20%5BDual%20Audio%5D%20%5BHindi%20(Line%20Audio)%20-%20English%5D%20ESubs%20-%20Downloadhub",
				BehaviorHints: &stremio.StreamBehaviorHints{
					BingeGroup: "torrentio|720p|BluRay|x264",
				},
			}, TorrentInfoInsertData{
				Hash:         "f5d0ab292f5a244a4b38efac9ae1f8d311179588",
				TorrentTitle: "Deadpool (2016) 720p BluRay x264 [Dual Audio] [Hindi (Line Audio) - English] ESubs - Downloadhub",
				Size:         -1,
				Source:       "tio",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := ExtractCreateDataFromStream(tc.hostname, tc.sid, &tc.stream)
			assert.Equal(t, &tc.data, data)
		})
	}
}

func TestExtractorMediaFusionTorrent(t *testing.T) {
	for _, tc := range []struct {
		name   string
		sid    string
		stream stremio.Stream
		data   TorrentInfoInsertData
	}{
		{
			"w/ tt - single",
			"tt1431045", stremio.Stream{
				Name:        "MediaFusion | ElfHosted P2P 2160P ⏳",
				Description: "📂 Deadpool 2016 4K HDR DV 2160p BDRemux Ita Eng x265-NAHOM\n💾 35.09 GB\n🌐 English + Italian\n🔗 Torlock",
				InfoHash:    "a3d11f4d97121a79f3e94b18a43e5b3e2f1853e1",
				FileIndex:   2,
				BehaviorHints: &stremio.StreamBehaviorHints{
					BingeGroup: "MediaFusion-|-ElfHosted-🎨 DV|HDR 📺 BluRay REMUX 🎞️ hevc-2160P",
					Filename:   "Deadpool.2016.4K.HDR.DV.2160p.BDRemux Ita Eng x265-NAHOM.mkv",
					VideoSize:  37682583137,
				},
			}, TorrentInfoInsertData{
				Hash:         "a3d11f4d97121a79f3e94b18a43e5b3e2f1853e1",
				TorrentTitle: "Deadpool 2016 4K HDR DV 2160p BDRemux Ita Eng x265-NAHOM",
				Size:         -1,
				Source:       "mfn",
				Files: []TorrentInfoInsertDataFile{
					{
						Name:   "Deadpool.2016.4K.HDR.DV.2160p.BDRemux Ita Eng x265-NAHOM.mkv",
						Idx:    2,
						Size:   37682583137,
						SId:    "tt1431045",
						Source: "",
					},
				},
			},
		},
		{
			"w/ tt - invalid behaviorHints.filename",
			"tt1431045", stremio.Stream{
				Name:        "MediaFusion | ElfHosted P2P 720P ⏳",
				Description: "📂 Deadpool 2016 x264 720p BluRay Eng Subs Dual Audio Hindi 5 1 English 5 1 Downloadhub\n💾 983.4 MB 👤 30\n🌐 English + Hindi\n🔗 TheRARBG",
				InfoHash:    "2decf5e42220711acf7a2515ed14ee78f13413fe",
				BehaviorHints: &stremio.StreamBehaviorHints{
					BingeGroup: "MediaFusion-|-ElfHosted-📺 BluRay 🎞️ avc-720P",
					Filename:   "Deadpool 2016 x264 720p BluRay Eng Subs Dual Audio Hindi 5 1 English 5 1 Downloadhub",
					VideoSize:  1031169664,
				},
			}, TorrentInfoInsertData{
				Hash:         "2decf5e42220711acf7a2515ed14ee78f13413fe",
				TorrentTitle: "Deadpool 2016 x264 720p BluRay Eng Subs Dual Audio Hindi 5 1 English 5 1 Downloadhub",
				Size:         -1,
				Source:       "mfn",
			},
		},
		{
			"w/ tt - multi",
			"tt1475582:1:1", stremio.Stream{
				Name:        "MediaFusion | ElfHosted P2P 1080P ⏳",
				Description: "📂 Sherlock S01-S04 + Extras Complete BluRay 1080p English DD 5 1 x264 ESub - mkvCinemas [Telly] ┈➤ Sherlock S01 E01 BluRay 1080p English DD 5 1 x264 ESub - mkvCinemas mkv\n💾 2.43 GB / 💾 33.48 GB\n🌐 English\n🔗 Zilean DMM",
				InfoHash:    "ce146cce125215f5e6615d2375ffa6a881c8eedd",
				FileIndex:   1,
				BehaviorHints: &stremio.StreamBehaviorHints{
					BingeGroup: "MediaFusion-|-ElfHosted-📺 BluRay 🎞️ avc 🎵 Dolby Digital-1080P",
					Filename:   "Sherlock S01 E01 BluRay 1080p English DD 5.1 x264 ESub - mkvCinemas.mkv",
					VideoSize:  2608894683,
				},
			}, TorrentInfoInsertData{
				Hash:         "ce146cce125215f5e6615d2375ffa6a881c8eedd",
				TorrentTitle: "Sherlock S01-S04 + Extras Complete BluRay 1080p English DD 5 1 x264 ESub - mkvCinemas [Telly]",
				Size:         35948876267,
				Source:       "mfn",
				Files: []TorrentInfoInsertDataFile{
					{
						Name:   "Sherlock S01 E01 BluRay 1080p English DD 5.1 x264 ESub - mkvCinemas.mkv",
						Idx:    1,
						Size:   2608894683,
						SId:    "tt1475582:1:1",
						Source: "",
					},
				},
			},
		},
		{
			"w/o tt - multi",
			"tt1475582:1:1", stremio.Stream{
				Name:        "MediaFusion | ElfHosted RD 1080P ⚡️",
				Description: "📺 BluRay 🎞️ avc 🎵 Dolby Digital\n💾 2.43 GB / 💾 33.48 GB\n🌐 English\n🔗 Zilean DMM",
				InfoHash:    "ce146cce125215f5e6615d2375ffa6a881c8eedd",
				FileIndex:   1,
				BehaviorHints: &stremio.StreamBehaviorHints{
					BingeGroup: "MediaFusion-|-ElfHosted-📺 BluRay 🎞️ avc 🎵 Dolby Digital-1080P",
					Filename:   "Sherlock S01 E01 BluRay 1080p English DD 5.1 x264 ESub - mkvCinemas.mkv",
					VideoSize:  2608894683,
				},
			}, TorrentInfoInsertData{
				Hash:         "ce146cce125215f5e6615d2375ffa6a881c8eedd",
				TorrentTitle: "",
				Size:         35948876267,
				Source:       "mfn",
				Files: []TorrentInfoInsertDataFile{
					{
						Name:   "Sherlock S01 E01 BluRay 1080p English DD 5.1 x264 ESub - mkvCinemas.mkv",
						Idx:    1,
						Size:   2608894683,
						SId:    "tt1475582:1:1",
						Source: "",
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := ExtractCreateDataFromStream("mediafusion.elfhosted.com", tc.sid, &tc.stream)
			assert.Equal(t, &tc.data, data)
		})
	}
}

func TestExtractorMediaFusionDebrid(t *testing.T) {
	for _, tc := range []struct {
		name   string
		sid    string
		stream stremio.Stream
		data   TorrentInfoInsertData
	}{
		{
			"w/ tt - single",
			"tt1431045", stremio.Stream{
				Name:        "MediaFusion | ElfHosted RD 2160P ⚡️",
				Description: "📂 Deadpool 2016 4K HDR DV 2160p BDRemux Ita Eng x265-NAHOM\n💾 35.09 GB\n🌐 English + Italian\n🔗 Torlock",
				URL:         "https://mediafusion.elfhosted.com/streaming_provider/D-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx-xxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx/stream/a3d11f4d97121a79f3e94b18a43e5b3e2f1853e1/Deadpool.2016.4K.HDR.DV.2160p.BDRemux%20Ita%20Eng%20x265-NAHOM.mkv",
				BehaviorHints: &stremio.StreamBehaviorHints{
					BingeGroup: "MediaFusion-|-ElfHosted-🎨 DV|HDR 📺 BluRay REMUX 🎞️ hevc-2160P",
					Filename:   "Deadpool.2016.4K.HDR.DV.2160p.BDRemux Ita Eng x265-NAHOM.mkv",
					VideoSize:  37682583137,
				},
			}, TorrentInfoInsertData{
				Hash:         "a3d11f4d97121a79f3e94b18a43e5b3e2f1853e1",
				TorrentTitle: "Deadpool 2016 4K HDR DV 2160p BDRemux Ita Eng x265-NAHOM",
				Size:         -1,
				Source:       "mfn",
				Files: []TorrentInfoInsertDataFile{
					{
						Name:   "Deadpool.2016.4K.HDR.DV.2160p.BDRemux Ita Eng x265-NAHOM.mkv",
						Idx:    -1,
						Size:   37682583137,
						SId:    "tt1431045",
						Source: "",
					},
				},
			},
		},
		{
			"w/ tt - invalid behaviorHints.filename",
			"tt1431045", stremio.Stream{
				Name:        "MediaFusion | ElfHosted RD 720P ⏳",
				Description: "📂 Deadpool 2016 x264 720p BluRay Eng Subs Dual Audio Hindi 5 1 English 5 1 Downloadhub\n💾 983.4 MB 👤 30\n🌐 English + Hindi\n🔗 TheRARBG",
				URL:         "https://mediafusion.elfhosted.com/streaming_provider/D-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx-xxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx/stream/2decf5e42220711acf7a2515ed14ee78f13413fe",
				BehaviorHints: &stremio.StreamBehaviorHints{
					BingeGroup: "MediaFusion-|-ElfHosted-📺 BluRay 🎞️ avc-720P",
					Filename:   "Deadpool 2016 x264 720p BluRay Eng Subs Dual Audio Hindi 5 1 English 5 1 Downloadhub",
					VideoSize:  1031169664,
				},
			}, TorrentInfoInsertData{
				Hash:         "2decf5e42220711acf7a2515ed14ee78f13413fe",
				TorrentTitle: "Deadpool 2016 x264 720p BluRay Eng Subs Dual Audio Hindi 5 1 English 5 1 Downloadhub",
				Size:         -1,
				Source:       "mfn",
			},
		},
		{
			"w/ tt - multi",
			"tt1475582:1:1", stremio.Stream{
				Name:        "MediaFusion | ElfHosted RD 1080P ⚡️",
				Description: "📂 Sherlock S01-S04 + Extras Complete BluRay 1080p English DD 5 1 x264 ESub - mkvCinemas [Telly] ┈➤ Sherlock S01 E01 BluRay 1080p English DD 5 1 x264 ESub - mkvCinemas mkv\n💾 2.43 GB / 💾 33.48 GB\n🌐 English\n🔗 Zilean DMM",
				URL:         "https://mediafusion.elfhosted.com/streaming_provider/D-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx-xxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx/stream/ce146cce125215f5e6615d2375ffa6a881c8eedd/1/1/Sherlock%20S01%20E01%20BluRay%201080p%20English%20DD%205.1%20x264%20ESub%20-%20mkvCinemas.mkv",
				BehaviorHints: &stremio.StreamBehaviorHints{
					BingeGroup: "MediaFusion-|-ElfHosted-📺 BluRay 🎞️ avc 🎵 Dolby Digital-1080P",
					Filename:   "Sherlock S01 E01 BluRay 1080p English DD 5.1 x264 ESub - mkvCinemas.mkv",
					VideoSize:  2608894683,
				},
			}, TorrentInfoInsertData{
				Hash:         "ce146cce125215f5e6615d2375ffa6a881c8eedd",
				TorrentTitle: "Sherlock S01-S04 + Extras Complete BluRay 1080p English DD 5 1 x264 ESub - mkvCinemas [Telly]",
				Size:         35948876267,
				Source:       "mfn",
				Files: []TorrentInfoInsertDataFile{
					{
						Name:   "Sherlock S01 E01 BluRay 1080p English DD 5.1 x264 ESub - mkvCinemas.mkv",
						Idx:    -1,
						Size:   2608894683,
						SId:    "tt1475582:1:1",
						Source: "",
					},
				},
			},
		},
		{
			"w/o tt - multi",
			"tt1475582:1:1", stremio.Stream{
				Name:        "MediaFusion | ElfHosted RD 1080P ⚡️",
				Description: "📺 BluRay 🎞️ avc 🎵 Dolby Digital\n💾 2.43 GB / 💾 33.48 GB\n🌐 English\n🔗 Zilean DMM",
				URL:         "https://mediafusion.elfhosted.com/streaming_provider/D-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx-xxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx/stream/ce146cce125215f5e6615d2375ffa6a881c8eedd/1/1/Sherlock%20S01%20E01%20BluRay%201080p%20English%20DD%205.1%20x264%20ESub%20-%20mkvCinemas.mkv",
				BehaviorHints: &stremio.StreamBehaviorHints{
					BingeGroup: "MediaFusion-|-ElfHosted-📺 BluRay 🎞️ avc 🎵 Dolby Digital-1080P",
					Filename:   "Sherlock S01 E01 BluRay 1080p English DD 5.1 x264 ESub - mkvCinemas.mkv",
					VideoSize:  2608894683,
				},
			}, TorrentInfoInsertData{
				Hash:         "ce146cce125215f5e6615d2375ffa6a881c8eedd",
				TorrentTitle: "",
				Size:         35948876267,
				Source:       "mfn",
				Files: []TorrentInfoInsertDataFile{
					{
						Name:   "Sherlock S01 E01 BluRay 1080p English DD 5.1 x264 ESub - mkvCinemas.mkv",
						Idx:    -1,
						Size:   2608894683,
						SId:    "tt1475582:1:1",
						Source: "",
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := ExtractCreateDataFromStream("mediafusion.elfhosted.com", tc.sid, &tc.stream)
			assert.Equal(t, &tc.data, data)
		})
	}
}
