package usecase

import (
	"context"
	"fmt"

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
		switch detail := cfg.Detail.(type) {
		case model.RedditFetchConfigDetail:
			data, err := uc.redditFetcher.Fetch(ctx, detail)
			if err != nil {
				// エラーログ
				continue
			}

			if err := uc.saveData(ctx, cfg.UserFetchConfig.ID, data); err != nil {
				// エラーログ
				continue
			}

		default:
			// 未対応のデータソースの場合
			continue
		}
	}
	return nil
}

func (uc *FetchAndSaveUsecase) saveData(ctx context.Context, configID uuid.UUID, data interface{}) error {
	return nil
}
