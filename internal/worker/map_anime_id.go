package worker

import (
	"slices"
	"strconv"
	"time"

	"github.com/MunifTanjim/stremthru/internal/anime"
	"github.com/MunifTanjim/stremthru/internal/anizip"
	"github.com/MunifTanjim/stremthru/internal/db"
	"github.com/MunifTanjim/stremthru/internal/logger"
	"github.com/MunifTanjim/stremthru/internal/util"
	"github.com/MunifTanjim/stremthru/internal/worker/worker_queue"
	"github.com/madflojo/tasks"
)

var anizipClient = anizip.NewAPIClient(&anizip.APIClientConfig{})

func InitMapAnimeIdWorker(conf *WorkerConfig) *Worker {
	if worker_queue.AnimeIdMapperQueue.Disabled {
		return nil
	}

	pool := anizip.GetMappingsPool()

	log := logger.Scoped("worker/map_anime_id")

	worker := &Worker{
		scheduler:  tasks.New(),
		shouldWait: conf.ShouldWait,
		onStart:    conf.OnStart,
		onEnd:      conf.OnEnd,
	}

	isRunning := false
	id, err := worker.scheduler.Add(&tasks.Task{
		Interval:          time.Duration(5 * time.Minute),
		RunSingleInstance: true,
		TaskFunc: func() (err error) {
			defer func() {
				if perr, stack := util.HandlePanic(recover(), true); perr != nil {
					err = perr
					log.Error("Worker Panic", "error", err, "stack", stack)
				} else {
					isRunning = false
				}
				worker.onEnd()
			}()

			for {
				wait, reason := worker.shouldWait()
				if !wait {
					break
				}
				log.Info("waiting, " + reason)
				time.Sleep(5 * time.Minute)
			}
			worker.onStart()

			if isRunning {
				return nil
			}

			isRunning = true

			worker_queue.AnimeIdMapperQueue.ProcessGroup(func(service string, items []worker_queue.AnimeIdMapperQueueItem) error {
				if service != anime.IdMapColumn.AniList {
					return nil
				}

				anilistIds := make([]int, len(items))
				for i := range items {
					id, err := strconv.Atoi(items[i].Id)
					if err != nil {
						return err
					}
					anilistIds[i] = id
				}

				idMaps, err := anime.GetIdMapsForAniList(anilistIds)
				if err != nil {
					return err
				}
				idMapByAnilistId := make(map[string]*anime.AnimeIdMap, len(idMaps))
				for i := range idMaps {
					idMap := &idMaps[i]
					idMapByAnilistId[idMap.AniList] = idMap
				}

				for cAnilistIds := range slices.Chunk(anilistIds, 100) {
					group := pool.NewGroup()

					for _, anilistId := range cAnilistIds {
						id := strconv.Itoa(anilistId)
						if idMap, ok := idMapByAnilistId[id]; !ok || idMap.IsStale() {
							if !ok {
								log.Debug("fetching missing idMap", "anilist_id", anilistId)
							} else {
								log.Debug("fetching stale idMap", "anilist_id", anilistId)
							}
							group.SubmitErr(func() (*anizip.GetMappingsData, error) {
								return anizipClient.GetMappings(&anizip.GetMappingsParams{
									Service: service,
									Id:      id,
								})
							})
						}
					}

					results, err := group.Wait()
					if err != nil {
						log.Error("failed to get mappings", "error", err, "service", service)
						return err
					}
					mapItems := []anime.AnimeIdMap{}
					for i := range results {
						m := results[i].Mappings
						mapItems = append(mapItems, anime.AnimeIdMap{
							Type:        m.Type,
							AniDB:       strconv.Itoa(m.AniDB),
							AniList:     strconv.Itoa(m.AniList),
							AniSearch:   strconv.Itoa(m.AniSearch),
							AnimePlanet: m.AnimePlanet,
							IMDB:        m.IMDB,
							Kitsu:       strconv.Itoa(m.Kitsu),
							LiveChart:   strconv.Itoa(m.LiveChart),
							MAL:         strconv.Itoa(m.MAL),
							NotifyMoe:   m.NotifyMoe,
							TMDB:        m.TMDB,
							TVDB:        strconv.Itoa(m.TVDB),
							UpdatedAt:   db.Timestamp{Time: time.Now()},
						})
					}
					err = anime.BulkRecordIdMaps(mapItems, service)
					if err != nil {
						log.Error("failed to record", "error", err)
						return err
					}
					time.Sleep(5 * time.Second)
				}

				return nil
			})

			return nil
		},
		ErrFunc: func(err error) {
			log.Error("Worker Failure", "error", err)

			isRunning = false
		},
	})

	if err != nil {
		panic(err)
	}

	log.Info("Started Worker", "id", id)

	return worker
}
