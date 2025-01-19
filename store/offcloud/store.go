package offcloud

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/MunifTanjim/stremthru/core"
	"github.com/MunifTanjim/stremthru/store"
)

type StoreClientConfig struct {
	HTTPClient *http.Client
	UserAgent  string
}

type StoreClient struct {
	Name   store.StoreName
	client *APIClient
}

func NewStoreClient(config *StoreClientConfig) *StoreClient {
	c := &StoreClient{}
	c.client = NewAPIClient(&APIClientConfig{
		HTTPClient: config.HTTPClient,
		UserAgent:  config.UserAgent,
	})
	c.Name = store.StoreNameOffcloud

	return c
}

func (s *StoreClient) GetName() store.StoreName {
	return s.Name
}

func (s *StoreClient) getMagnetFiles(ctx Ctx, requestId string, magnetName string) ([]store.MagnetFile, error) {
	files := []store.MagnetFile{}
	res, err := s.client.ExploreCloudDownload(&ExploreCloudDownloadParams{
		Ctx:       ctx,
		RequestId: requestId,
	})
	if err != nil {
		return files, err
	}
	for _, link := range res.Data {
		info, err := link.parse()
		if err != nil {
			return nil, err
		}
		size := -1
		// // too expensive, should enable for non-public deployments later 
		// if size_res, err := s.client.GetFileSize(&GetFileSizeParams{Ctx: ctx, Link: string(link)}); err == nil {
		// 	size = size_res.Data
		// }
		file := store.MagnetFile{
			Idx:  info.fileIdx,
			Link: string(link),
			Name: info.fileName,
			Path: "/" + magnetName + "/" + filepath.Base(info.fileName),
			Size: size,
		}
		files = append(files, file)
	}
	return files, nil
}

func (s *StoreClient) AddMagnet(params *store.AddMagnetParams) (*store.AddMagnetData, error) {
	magnet, err := core.ParseMagnetLink(params.Magnet)
	if err != nil {
		return nil, err
	}
	lm_params := &store.ListMagnetsParams{
		Ctx: params.Ctx,
	}
	lm_res, err := s.ListMagnets(lm_params)
	if err != nil {
		return nil, err
	}

	var lmi *store.ListMagnetsDataItem
	for i := range lm_res.Items {
		item := &lm_res.Items[i]
		if item.Hash == magnet.Hash {
			lmi = item
			break
		}
	}

	data := &store.AddMagnetData{
		Hash:   magnet.Hash,
		Magnet: magnet.Link,
		Files:  []store.MagnetFile{},
	}

	if lmi != nil {
		data.Id = lmi.Id
		data.Name = lmi.Name
		data.Status = lmi.Status
		data.AddedAt = lmi.AddedAt
	} else {
		res, err := s.client.AddCloudDownload(&AddCloudDownloadParams{
			Ctx: params.Ctx,
			URL: magnet.Link,
		})
		if err != nil {
			return nil, err
		}

		data.Id = res.Data.RequestId
		data.Name = res.Data.FileName
		data.Status = getMagnetStatus(res.Data.Status)
		data.AddedAt = res.Data.CreatedOn
	}

	if data.Status == store.MagnetStatusDownloaded {
		files, err := s.getMagnetFiles(params.Ctx, data.Id, data.Name)
		if err != nil {
			return nil, err
		}
		data.Files = files
	}
	return data, nil
}

func (s *StoreClient) CheckMagnet(params *store.CheckMagnetParams) (*store.CheckMagnetData, error) {
	hashes := []string{}
	magnetByHash := map[string]core.MagnetLink{}
	for _, magnet := range params.Magnets {
		if m, err := core.ParseMagnetLink(magnet); err == nil {
			hashes = append(hashes, m.Hash)
			magnetByHash[m.Hash] = m
		}
	}
	res, err := s.client.CheckCache(&CheckCacheParams{
		Ctx:    params.Ctx,
		Hashes: hashes,
	})
	if err != nil {
		return nil, err
	}
	data := &store.CheckMagnetData{
		Items: []store.CheckMagnetDataItem{},
	}
	cachedByHash := map[string]bool{}
	for _, hash := range res.Data.CachedItems {
		cachedByHash[hash] = true
	}
	for _, hash := range hashes {
		m := magnetByHash[hash]
		item := store.CheckMagnetDataItem{
			Hash:   m.Hash,
			Magnet: m.Link,
			Status: store.MagnetStatusUnknown,
			Files:  []store.MagnetFile{},
		}
		if _, ok := cachedByHash[m.Hash]; ok {
			item.Status = store.MagnetStatusCached
		}
		data.Items = append(data.Items, item)
	}
	return data, nil
}

func (s *StoreClient) GenerateLink(params *store.GenerateLinkParams) (*store.GenerateLinkData, error) {
	data := &store.GenerateLinkData{Link: params.Link}
	return data, nil
}

func getMagnetStatus(status CloudDownloadStatus) store.MagnetStatus {
	switch status {
	case CloudDownloadStatusCreated:
		return store.MagnetStatusDownloading
	case CloudDownloadStatusDownloaded:
		return store.MagnetStatusDownloaded
	case CloudDownloadStatusError:
		return store.MagnetStatusFailed
	default:
		return store.MagnetStatusUnknown
	}
}

func (s *StoreClient) GetMagnet(params *store.GetMagnetParams) (*store.GetMagnetData, error) {
	res, err := s.client.GetCloudDownloadStatus(&GetCloudDownloadStatusParams{
		Ctx:       params.Ctx,
		RequestId: params.Id,
	})
	if err != nil {
		return nil, err
	}
	magnet := res.Data.Status
	data := &store.GetMagnetData{
		Id:      params.Id,
		Name:    magnet.FileName,
		Hash:    "",
		Status:  getMagnetStatus(magnet.Status),
		Files:   []store.MagnetFile{},
		AddedAt: time.Now(),
	}
	if data.Status == store.MagnetStatusDownloaded {
		files, err := s.getMagnetFiles(params.Ctx, data.Id, data.Name)
		if err != nil {
			return nil, err
		}
		data.Files = files
	}
	return data, nil
}

func (s *StoreClient) GetUser(params *store.GetUserParams) (*store.User, error) {
	email_res, err := s.client.GetUserEmail(&GetUserEmailParams{
		Ctx: params.Ctx,
	})
	if err != nil {
		return nil, err
	}
	data := &store.User{
		Id:                 email_res.Data.UserId,
		Email:              email_res.Data.Email,
		SubscriptionStatus: store.UserSubscriptionStatusTrial,
	}
	stats_res, err := s.client.GetAccountStats(&GetAccountStatsParams{
		Ctx: params.Ctx,
	})
	if err != nil {
		return nil, err
	}
	if stats_res.Data.ExpirationDate.After(time.Now()) {
		data.SubscriptionStatus = store.UserSubscriptionStatusPremium
	}
	return data, nil
}

func (s *StoreClient) ListMagnets(params *store.ListMagnetsParams) (*store.ListMagnetsData, error) {
	res, err := s.client.ListCloudDownloads(&ListCloudDownloadsParams{
		Ctx: params.Ctx,
	})
	if err != nil {
		return nil, err
	}
	data := &store.ListMagnetsData{
		Items:      []store.ListMagnetsDataItem{},
		TotalItems: len(res.Data.History),
	}
	for _, m := range res.Data.History {
		magnet, err := core.ParseMagnetLink(m.OriginalLink)
		if err != nil {
			continue
		}
		item := store.ListMagnetsDataItem{
			Id:      m.RequestId,
			Hash:    magnet.Hash,
			Name:    m.FileName,
			Status:  getMagnetStatus(m.Status),
			AddedAt: m.CreatedOn,
		}
		data.Items = append(data.Items, item)
	}
	return data, nil
}

func (s *StoreClient) RemoveMagnet(params *store.RemoveMagnetParams) (*store.RemoveMagnetData, error) {
	_, err := s.client.RemoveCloudDownload(&RemoveCloudDownloadParams{
		Ctx:       params.Ctx,
		RequestId: params.Id,
	})
	if err != nil {
		return nil, err
	}
	data := &store.RemoveMagnetData{Id: params.Id}
	return data, nil
}
