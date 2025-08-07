package di

import (
	"fmt"

	"github.com/YamaguchiKoki/feedle_batch/internal/adapter/fetcher"
	"github.com/YamaguchiKoki/feedle_batch/internal/adapter/fetcher/reddit"
	"github.com/YamaguchiKoki/feedle_batch/internal/adapter/repository"
	"github.com/YamaguchiKoki/feedle_batch/internal/domain/model"
	"github.com/YamaguchiKoki/feedle_batch/internal/domain/service"
	"github.com/YamaguchiKoki/feedle_batch/internal/port/output"
	"github.com/YamaguchiKoki/feedle_batch/internal/usecase"
	"github.com/samber/do"
	"github.com/spf13/viper"
	"github.com/supabase-community/supabase-go"
)

func NewContainer() (*do.Injector, error) {
	injector := do.New()

	// Register Supabase client
	do.Provide(injector, func(i *do.Injector) (*supabase.Client, error) {
		supabaseURL := viper.GetString("SUPABASE_URL")
		supabaseKey := viper.GetString("SUPABASE_ANON_KEY")
		if supabaseURL == "" || supabaseKey == "" {
			return nil, fmt.Errorf("SUPABASE_URL and SUPABASE_ANON_KEY must be set")
		}

		return supabase.NewClient(supabaseURL, supabaseKey, nil)
	})

	// Register repositories
	do.Provide(injector, func(i *do.Injector) (output.UserRepository, error) {
		client := do.MustInvoke[*supabase.Client](i)
		return repository.NewSupabaseUserRepository(client), nil
	})

	do.Provide(injector, func(i *do.Injector) (output.FetchConfigRepository, error) {
		client := do.MustInvoke[*supabase.Client](i)
		return repository.NewSupabaseFetchConfigRepository(client), nil
	})

	do.Provide(injector, func(i *do.Injector) (output.FetchedDataRepository, error) {
		client := do.MustInvoke[*supabase.Client](i)
		return repository.NewSupabaseFetchedDataRepository(client), nil
	})

	do.Provide(injector, func(i *do.Injector) (output.DataSourceRepository, error) {
		client := do.MustInvoke[*supabase.Client](i)
		return repository.NewSupabaseDataSourceRepository(client), nil
	})

	do.Provide(injector, func(i *do.Injector) (output.RedditFetchConfigRepository, error) {
		client := do.MustInvoke[*supabase.Client](i)
		return repository.NewSupabaseRedditFetchConfigRepository(client), nil
	})

	// Register services
	do.Provide(injector, func(i *do.Injector) (*service.FetchConfigService, error) {
		userRepo := do.MustInvoke[output.UserRepository](i)
		configRepo := do.MustInvoke[output.FetchConfigRepository](i)
		redditConfigRepo := do.MustInvoke[output.RedditFetchConfigRepository](i)

		return service.NewFetchConfigService(
			userRepo,
			configRepo,
			redditConfigRepo,
		), nil
	})

	// Register fetchers
	do.Provide(injector, func(i *do.Injector) (fetcher.Fetcher[model.RedditFetchConfigDetail], error) {
		redditClientID := viper.GetString("REDDIT_CLIENT_ID")
		redditClientSecret := viper.GetString("REDDIT_CLIENT_SECRET")
		redditUsername := viper.GetString("REDDIT_USERNAME")

		return reddit.NewRedditFetcher(
			nil,
			redditClientID,
			redditClientSecret,
			redditUsername,
		), nil
	})

	// Register usecase
	do.Provide(injector, func(i *do.Injector) (*usecase.FetchAndSaveUsecase, error) {
		fetchConfigService := do.MustInvoke[*service.FetchConfigService](i)
		dataRepo := do.MustInvoke[output.FetchedDataRepository](i)
		redditFetcher := do.MustInvoke[fetcher.Fetcher[model.RedditFetchConfigDetail]](i)

		return usecase.NewFetchAndSaveUsecase(
			fetchConfigService,
			dataRepo,
			redditFetcher,
		), nil
	})

	return injector, nil
}
