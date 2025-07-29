package di

import (
	"fmt"

	"github.com/YamaguchiKoki/feedle_batch/internal/config"
	"github.com/YamaguchiKoki/feedle_batch/internal/fetcher"
	"github.com/YamaguchiKoki/feedle_batch/internal/fetcher/reddit"
	"github.com/YamaguchiKoki/feedle_batch/internal/repository"
	"github.com/YamaguchiKoki/feedle_batch/internal/service"
	"github.com/YamaguchiKoki/feedle_batch/internal/usecase"
	"github.com/samber/do/v2"
	"github.com/supabase-community/supabase-go"
)

// ProvideConfig provides the application configuration
func ProvideConfig(i do.Injector) (*config.Config, error) {
	return config.Load(), nil
}

// ProvideSupabaseClient provides the Supabase client
func ProvideSupabaseClient(i do.Injector) (*supabase.Client, error) {
	cfg := do.MustInvoke[*config.Config](i)

	if cfg.SupabaseURL == "" || cfg.SupabaseServiceKey == "" {
		return nil, fmt.Errorf("SUPABASE_URL and SUPABASE_SERVICE_KEY are required")
	}

	client, err := supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseServiceKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Supabase client: %w", err)
	}

	return client, nil
}

// ProvideFetchedDataRepository provides the appropriate repository based on dry-run mode
func ProvideFetchedDataRepository(i do.Injector) (repository.FetchedDataRepository, error) {
	// Check if we're in dry-run mode via named service
	isDryRun, err := do.InvokeNamed[bool](i, "dry-run")
	if err != nil {
		// Default to false if not set
		isDryRun = false
	}

	if isDryRun {
		return repository.NewMockFetchedDataRepository(), nil
	}

	client := do.MustInvoke[*supabase.Client](i)
	return repository.NewSupabaseFetchedDataRepository(client), nil
}

// ProvideDataService provides the data service
func ProvideDataService(i do.Injector) (*service.DataService, error) {
	repo := do.MustInvoke[repository.FetchedDataRepository](i)
	return service.NewDataService(repo), nil
}

// ProvideFetcherRegistry provides the fetcher registry with all fetchers registered
func ProvideFetcherRegistry(i do.Injector) (*fetcher.Registry, error) {
	registry := fetcher.NewRegistry()

	// Register Reddit fetcher
	redditFetcher := do.MustInvoke[*reddit.RedditFetcher](i)
	registry.Register(redditFetcher)

	// Future: Register other fetchers
	// twitterFetcher := do.MustInvoke[*twitter.TwitterFetcher](i)
	// registry.Register(twitterFetcher)

	return registry, nil
}

// ProvideRedditFetcher provides the Reddit fetcher
func ProvideRedditFetcher(i do.Injector) (*reddit.RedditFetcher, error) {
	cfg := do.MustInvoke[*config.Config](i)

	return reddit.NewRedditFetcher(
		nil, // HTTP client (nil uses default)
		cfg.RedditClientID,
		cfg.RedditClientSecret,
		cfg.RedditUsername,
	), nil
}

// ProvideFetchUseCase provides the fetch use case
func ProvideFetchUseCase(i do.Injector) (*usecase.FetchUseCase, error) {
	dataService := do.MustInvoke[*service.DataService](i)
	fetcherRegistry := do.MustInvoke[*fetcher.Registry](i)
	return usecase.NewFetchUseCase(dataService, fetcherRegistry), nil
}

// RegisterServices registers all services with the DI container
func RegisterServices(i do.Injector, isDryRun bool) error {
	// Register dry-run flag as a named service
	do.ProvideNamedValue(i, "dry-run", isDryRun)

	// Register all providers
	do.Provide(i, ProvideConfig)

	if !isDryRun {
		do.Provide(i, ProvideSupabaseClient)
	}

	do.Provide(i, ProvideFetchedDataRepository)
	do.Provide(i, ProvideDataService)
	do.Provide(i, ProvideRedditFetcher)
	do.Provide(i, ProvideFetcherRegistry)
	do.Provide(i, ProvideFetchUseCase)

	return nil
}

// RegisterTestServices registers services for testing with mocks
func RegisterTestServices(i do.Injector) error {
	// Always use mock repository for tests
	do.ProvideNamedValue(i, "dry-run", true)

	// Register test config
	do.ProvideValue(i, &config.Config{
		RedditClientID:     "test-client-id",
		RedditClientSecret: "test-client-secret",
		RedditUsername:     "test-username",
	})

	// Register all providers
	do.Provide(i, ProvideFetchedDataRepository)
	do.Provide(i, ProvideDataService)
	do.Provide(i, ProvideRedditFetcher)
	do.Provide(i, ProvideFetcherRegistry)
	do.Provide(i, ProvideFetchUseCase)

	return nil
}
