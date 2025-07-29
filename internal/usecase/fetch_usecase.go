package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/YamaguchiKoki/feedle_batch/internal/fetcher"
	"github.com/YamaguchiKoki/feedle_batch/internal/models"
	"github.com/YamaguchiKoki/feedle_batch/internal/service"
	"github.com/samber/lo"
)

type FetchUseCase struct {
	dataService     *service.DataService
	fetcherRegistry *fetcher.Registry
}

func NewFetchUseCase(dataService *service.DataService, fetcherRegistry *fetcher.Registry) *FetchUseCase {
	return &FetchUseCase{
		dataService:     dataService,
		fetcherRegistry: fetcherRegistry,
	}
}

type FetchOptions struct {
	Sources    []string
	ConfigID   string
	DryRun     bool
	Limit      int
	Keywords   []string
	Subreddits []string
}

type FetchResult struct {
	Source      string
	TotalFound  int
	SaveResult  *service.SaveResult
	Error       error
}

func (uc *FetchUseCase) FetchFromSources(ctx context.Context, opts FetchOptions) ([]FetchResult, error) {
	// Determine sources to fetch from
	sourcesToFetch := opts.Sources
	if len(sourcesToFetch) == 0 {
		// Default to all registered sources
		sourcesToFetch = lo.Map(uc.fetcherRegistry.GetAll(), func(f fetcher.Fetcher, _ int) string {
			return f.Name()
		})
	}

	results := make([]FetchResult, 0, len(sourcesToFetch))
	
	// TODO: Parallelize this with goroutines
	for _, source := range sourcesToFetch {
		result := uc.fetchFromSource(ctx, source, opts)
		results = append(results, result)
		
		if result.Error != nil {
			log.Printf("Error fetching from %s: %v\n", source, result.Error)
		}
	}
	
	return results, nil
}

func (uc *FetchUseCase) fetchFromSource(ctx context.Context, sourceName string, opts FetchOptions) FetchResult {
	result := FetchResult{
		Source: sourceName,
	}

	f, ok := uc.fetcherRegistry.Get(sourceName)
	if !ok {
		result.Error = fmt.Errorf("unknown source: %s", sourceName)
		return result
	}

	fmt.Printf("Fetching from %s...\n", sourceName)

	// Build fetch configuration
	config := uc.buildFetchConfig(sourceName, opts)

	// Fetch data
	posts, err := f.Fetch(config)
	if err != nil {
		result.Error = fmt.Errorf("failed to fetch: %w", err)
		return result
	}

	result.TotalFound = len(posts)
	fmt.Printf("Found %d posts\n", len(posts))

	if opts.DryRun {
		// Dry-run mode: display first 3 posts
		uc.displayDryRunResults(posts)
		result.SaveResult = &service.SaveResult{
			Total: len(posts),
			Saved: 0,
		}
	} else {
		// Save to database
		saveResult, err := uc.saveToDatabase(ctx, posts, opts.ConfigID)
		if err != nil {
			result.Error = fmt.Errorf("failed to save data: %w", err)
			return result
		}
		result.SaveResult = saveResult
		
		fmt.Printf("Save result for %s: %s\n", sourceName, saveResult.Summary())
		
		// Display errors if any
		if len(saveResult.Errors) > 0 {
			fmt.Println("Errors encountered:")
			for i, errMsg := range saveResult.Errors {
				if i >= 5 { // Max 5 errors displayed
					fmt.Printf("... and %d more errors\n", len(saveResult.Errors)-5)
					break
				}
				fmt.Printf("  - %s\n", errMsg)
			}
		}
	}

	return result
}

func (uc *FetchUseCase) buildFetchConfig(sourceName string, opts FetchOptions) fetcher.FetchConfig {
	config := fetcher.FetchConfig{
		Keywords: opts.Keywords,
		Limit:    opts.Limit,
	}

	// Source-specific configuration
	switch sourceName {
	case "reddit":
		if len(opts.Subreddits) == 0 {
			opts.Subreddits = []string{"golang", "programming"}
		}
		config.Reddit.Subreddits = opts.Subreddits
	// Future sources
	// case "youtube":
	//     config.YouTube.ChannelIDs = opts.ChannelIDs
	}

	return config
}

func (uc *FetchUseCase) displayDryRunResults(posts []*models.FetchedData) {
	for i, post := range lo.Slice(posts, 0, 3) {
		fmt.Printf("  [%d] %s (score: %v)\n",
			i+1,
			post.Title,
			post.Engagement["score"])
	}
}

func (uc *FetchUseCase) saveToDatabase(ctx context.Context, posts []*models.FetchedData, configID string) (*service.SaveResult, error) {
	if len(posts) == 0 {
		return &service.SaveResult{}, nil
	}

	opts := service.SaveOptions{
		ConfigID:            configID,
		SkipDuplicatesByURL: true,
		BatchSize:           50,
	}

	return uc.dataService.SaveFetchedData(ctx, posts, opts)
}