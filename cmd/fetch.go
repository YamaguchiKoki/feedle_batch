package cmd

import (
	"context"
	"log"
	"time"

	"github.com/YamaguchiKoki/feedle_batch/internal/di"
	"github.com/YamaguchiKoki/feedle_batch/internal/usecase"
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
)

var (
	sources    []string
	subreddits []string
	keywords   []string
	limit      int
	dryRun     bool
	configID   string
)


var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch data from configured sources",
	Long:  `Fetch data from various sources (Reddit, Twitter, etc.) and save to database`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create DI container
		injector := do.New()
		defer func() {
			if err := injector.Shutdown(); err != nil {
				log.Printf("Warning: Failed to shutdown injector: %v", err)
			}
		}()

		// Register services
		if err := di.RegisterServices(injector, dryRun); err != nil {
			log.Fatal("Failed to register services:", err)
		}

		// Get fetch use case
		fetchUseCase, err := do.Invoke[*usecase.FetchUseCase](injector)
		if err != nil {
			log.Fatal("Failed to get fetch use case:", err)
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		// Execute fetch
		opts := usecase.FetchOptions{
			Sources:    sources,
			ConfigID:   configID,
			DryRun:     dryRun,
			Limit:      limit,
			Keywords:   keywords,
			Subreddits: subreddits,
		}

		results, err := fetchUseCase.FetchFromSources(ctx, opts)
		if err != nil {
			log.Fatal("Failed to fetch data:", err)
		}

		// Summary
		totalSaved := 0
		totalErrors := 0
		for _, result := range results {
			if result.SaveResult != nil {
				totalSaved += result.SaveResult.Saved
			}
			if result.Error != nil {
				totalErrors++
			}
		}

		if !dryRun {
			log.Printf("Fetch completed. Total saved: %d, Sources with errors: %d", totalSaved, totalErrors)
		}
	},
}


func init() {
	rootCmd.AddCommand(fetchCmd)

	fetchCmd.Flags().StringSliceVarP(&sources, "sources", "s", []string{}, "Sources to fetch from (default: all)")
	fetchCmd.Flags().StringSliceVar(&subreddits, "subreddits", []string{}, "Reddit subreddits to fetch")
	fetchCmd.Flags().StringSliceVar(&keywords, "keywords", []string{}, "Keywords to search")
	fetchCmd.Flags().IntVar(&limit, "limit", 25, "Maximum number of items to fetch per source")
	fetchCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Run without saving to database")
	fetchCmd.Flags().StringVar(&configID, "config-id", "default", "Configuration ID for data tracking")
}
