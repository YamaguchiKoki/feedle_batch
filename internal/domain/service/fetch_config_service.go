package service

import (
	"context"
	"fmt"

	"github.com/YamaguchiKoki/feedle_batch/internal/domain/model"
	"github.com/YamaguchiKoki/feedle_batch/internal/port/output"
)

type FetchConfigService struct {
	userRepo              output.UserRepository
	configRepo            output.FetchConfigRepository
	redditFetchConfigRepo output.RedditFetchConfigRepository
}

type EnrichedFetchConfig struct {
	UserFetchConfig model.UserFetchConfig
	Detail          model.FetchConfigDetail
}

func NewFetchConfigService(
	uRepo output.UserRepository,
	cRepo output.FetchConfigRepository,
	rRepo output.RedditFetchConfigRepository,
) *FetchConfigService {
	return &FetchConfigService{
		userRepo:              uRepo,
		configRepo:            cRepo,
		redditFetchConfigRepo: rRepo,
	}
}

// 全てのアクティブユーザーの取得設定を取得する
func (s *FetchConfigService) GetActiveUsersEnrichedConfigs(ctx context.Context) ([]EnrichedFetchConfig, error) {
	activeUserIDs, err := s.userRepo.GetActiveUserIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}

	var allConfigs []EnrichedFetchConfig

	for _, userID := range activeUserIDs {
		userConfigs, err := s.GetUserEnrichedConfigs(ctx, userID)
		if err != nil {
			// エラーログを記録して続行
			continue
		}
		allConfigs = append(allConfigs, userConfigs...)
	}

	return allConfigs, nil
}

// 特定ユーザーの設定を詳細情報付きで取得
func (s *FetchConfigService) GetUserEnrichedConfigs(ctx context.Context, userID model.UserID) ([]EnrichedFetchConfig, error) {
	configs, err := s.configRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get configs for user %s: %w", userID, err)
	}

	enrichedConfigs := make([]EnrichedFetchConfig, 0, len(configs))

	for _, config := range configs {
		if !config.IsActive {
			continue
		}

		detail, err := s.getConfigDetail(ctx, *config)
		if err != nil {
			// エラーログを記録して続行
			continue
		}

		enrichedConfigs = append(enrichedConfigs, EnrichedFetchConfig{
			UserFetchConfig: *config,
			Detail:          detail,
		})
	}

	return enrichedConfigs, nil
}

// データソースに応じて適切な詳細設定を取得
func (s *FetchConfigService) getConfigDetail(ctx context.Context, config model.UserFetchConfig) (model.FetchConfigDetail, error) {
	switch config.DataSourceID {
	case "reddit":
		return s.redditFetchConfigRepo.GetByUserFetchConfigID(ctx, config.ID)
	default:
		return nil, fmt.Errorf("unsupported data source: %s", config.DataSourceID)
	}
}
