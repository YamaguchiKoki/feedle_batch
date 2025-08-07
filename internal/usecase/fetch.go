package usecase

import (
	"context"
	"fmt"

	"github.com/YamaguchiKoki/feedle_batch/internal/adapter/fetcher"
	"github.com/YamaguchiKoki/feedle_batch/internal/port/output"
)

type FetchAndSaveUsecase struct {
	userRepository              output.UserRepository
	configRepository            output.FetchConfigRepository
	dataRepository              output.FetchedDataRepository
	dataSourceRepository        output.DataSourceRepository
	redditFetchConfigRepository output.RedditFetchConfigRepository
	fetcherRegistry             fetcher.Registry
}

func NewFetchAndSaveUsecase(
	uRepo output.UserRepository,
	cRepo output.FetchConfigRepository,
	dRepo output.FetchedDataRepository,
	dsRepo output.DataSourceRepository,
	rRepo output.RedditFetchConfigRepository,
	r fetcher.Registry) *FetchAndSaveUsecase {
	return &FetchAndSaveUsecase{
		userRepository:              uRepo,
		configRepository:            cRepo,
		dataRepository:              dRepo,
		dataSourceRepository:        dsRepo,
		redditFetchConfigRepository: rRepo,
		fetcherRegistry:             r,
	}
}

func (uc *FetchAndSaveUsecase) Execute(ctx context.Context) error {
	activeUserIDs, err := uc.userRepository.GetActiveUserIDs(ctx)
	if err != nil {
		fmt.Errorf("Fail: userRepository.GetActiveUserIDs(ctx)")
	}
	for _, id := range activeUserIDs {
		userFetchConfigs, err := uc.configRepository.GetByUserID(ctx, id)
		if err != nil {
			fmt.Errorf("Fail: configRepository.GetByUserID(ctx, id)")
		}

		for _, config := range userFetchConfigs {
			switch config.DataSourceID {
			case "reddit":
				redditConfig, err := uc.redditFetchConfigRepository.GetByUserFetchConfigID(ctx, config.ID)
				if err != nil {
					fmt.Errorf("Fail: redditFetchConfigRepository.GetByUserFetchConfigID: %v", err, redditConfig)
					continue
				}
			default:
				fmt.Errorf("unsupported data source: %s", config.DataSourceID)
			}

		}
	}
}
