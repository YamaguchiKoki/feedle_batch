package cmd

import (
	"context"
	"log"

	"github.com/YamaguchiKoki/feedle_batch/internal/adapter/fetcher/reddit"
	"github.com/YamaguchiKoki/feedle_batch/internal/adapter/repository"
	"github.com/YamaguchiKoki/feedle_batch/internal/domain/service"
	"github.com/YamaguchiKoki/feedle_batch/internal/usecase"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/supabase-community/supabase-go"
)

var (
	dryRun bool
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch data from configured sources",
	Long:  `Fetch data from various sources (Reddit, Twitter, etc.) and save to database`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// Initialize Supabase client
		supabaseURL := viper.GetString("SUPABASE_URL")
		supabaseKey := viper.GetString("SUPABASE_ANON_KEY")
		if supabaseURL == "" || supabaseKey == "" {
			log.Fatal("SUPABASE_URL and SUPABASE_ANON_KEY must be set")
		}

		supabaseClient, err := supabase.NewClient(supabaseURL, supabaseKey, nil)
		if err != nil {
			log.Fatal("Failed to create Supabase client:", err)
		}

		// Initialize repositories
		userRepo := repository.NewSupabaseUserRepository(supabaseClient)
		configRepo := repository.NewSupabaseFetchConfigRepository(supabaseClient)
		dataRepo := repository.NewSupabaseFetchedDataRepository(supabaseClient)
		// dataSourceRepo := repository.NewSupabaseDataSourceRepository(supabaseClient)
		redditConfigRepo := repository.NewSupabaseRedditFetchConfigRepository(supabaseClient)

		// Initialize service
		fetchConfigService := service.NewFetchConfigService(
			userRepo,
			configRepo,
			redditConfigRepo,
		)

		// Initialize fetchers
		redditClientID := viper.GetString("REDDIT_CLIENT_ID")
		redditClientSecret := viper.GetString("REDDIT_CLIENT_SECRET")
		redditUsername := viper.GetString("REDDIT_USERNAME")

		redditFetcher := reddit.NewRedditFetcher(
			nil,
			redditClientID,
			redditClientSecret,
			redditUsername,
		)

		// Initialize usecase
		fetchUsecase := usecase.NewFetchAndSaveUsecase(
			fetchConfigService,
			dataRepo,
			redditFetcher,
		)

		// Execute fetch
		if err := fetchUsecase.Execute(ctx); err != nil {
			log.Fatal("Failed to execute fetch:", err)
		}

		log.Println("Fetch completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)

	fetchCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Run without saving to database")
}
