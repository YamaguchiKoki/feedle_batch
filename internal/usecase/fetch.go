package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/YamaguchiKoki/feedle_batch/internal/adapter/fetcher"
	"github.com/YamaguchiKoki/feedle_batch/internal/domain/model"
	"github.com/YamaguchiKoki/feedle_batch/internal/domain/service"
	"github.com/YamaguchiKoki/feedle_batch/internal/port/output"
	"github.com/google/uuid"
)

type FetchAndSaveUsecase struct {
	fetchConfigService *service.FetchConfigService
	dataRepository     output.FetchedDataRepository
	redditFetcher      fetcher.Fetcher[model.RedditFetchConfigDetail]
}

func NewFetchAndSaveUsecase(
	fetchConfigService *service.FetchConfigService,
	dRepo output.FetchedDataRepository,
	redditFetcher fetcher.Fetcher[model.RedditFetchConfigDetail],
) *FetchAndSaveUsecase {
	return &FetchAndSaveUsecase{
		fetchConfigService: fetchConfigService,
		dataRepository:     dRepo,
		redditFetcher:      redditFetcher,
	}
}

func (uc *FetchAndSaveUsecase) Execute(ctx context.Context) error {
	enrichedConfigs, err := uc.fetchConfigService.GetActiveUsersEnrichedConfigs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get enriched configs: %w", err)
	}

	for _, cfg := range enrichedConfigs {
		data, err := uc.fetchData(ctx, cfg)
		if err != nil {
			log.Printf("Failed to fetch data for config %s: %v", cfg.UserFetchConfig.ID, err)
			continue
		}

		if err := uc.saveData(ctx, cfg.UserFetchConfig.ID, data); err != nil {
			log.Printf("Failed to save data for config %s: %v", cfg.UserFetchConfig.ID, err)
			continue
		}

		log.Printf("Successfully processed config %s: fetched and saved %d items",
			cfg.UserFetchConfig.ID, len(data))
	}
	return nil
}

func (uc *FetchAndSaveUsecase) fetchData(ctx context.Context, cfg service.EnrichedFetchConfig) ([]*model.FetchedData, error) {
	switch detail := cfg.Detail.(type) {
	case model.RedditFetchConfigDetail:
		return uc.redditFetcher.Fetch(ctx, detail)
	default:
		return nil, fmt.Errorf("unsupported data source: %s", cfg.UserFetchConfig.DataSourceID)
	}
}

func (uc *FetchAndSaveUsecase) saveData(ctx context.Context, configID uuid.UUID, data []*model.FetchedData) error {
	for _, item := range data {
		// Set the config ID for each item
		// item.ConfigID = configID

		if err := uc.dataRepository.Create(ctx, item); err != nil {
			return fmt.Errorf("failed to save data item: %w", err)
		}
	}

	log.Printf("Saved %d items for config %s", len(data), configID)
	return nil
}
