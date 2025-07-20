package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/YamaguchiKoki/feedle_batch/internal/config"
	"github.com/YamaguchiKoki/feedle_batch/internal/fetcher"
	"github.com/YamaguchiKoki/feedle_batch/internal/fetcher/reddit"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/supabase-community/supabase-go"
)

var (
	sources    []string
	subreddits []string
	keywords   []string
	limit      int
	dryRun     bool
)

var fetcherRegistry *fetcher.Registry

var fetchCmd = &cobra.Command{
	Use: "fetch",
	PreRun: func(cmd *cobra.Command, args []string) {
		initializeFetchers()
	},
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()

		if !dryRun && (cfg.SupabaseServiceKey == "" || cfg.SupabaseURL == "") {
			log.Fatal("SUPABASE_URL and SUPABASE_SERVICE_KEY are required")
		}

		var supabaseClient *supabase.Client
		if !dryRun {
			client, err := supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseURL, nil)
			if err != nil {
				log.Fatal("Failed to create Supabase client:", err)
			}
			supabaseClient = client
		}

		// 実行するソースを決定
		sourcesToFetch := sources
		if len(sourcesToFetch) == 0 {
			// デフォルトは全ソース
			sourcesToFetch = lo.Map(fetcherRegistry.GetAll(), func(f fetcher.Fetcher, _ int) string {
				return f.Name()
			})
		}

		// TODO: 並行化
		for _, source := range sourcesToFetch {
			if err := fetchFromSource(source, supabaseClient); err != nil {
				log.Printf("Error fetching from %s: %v\n", source, err)
			}
		}
	},
}

func fetchFromSource(sourceName string, client *supabase.Client) error { //nolint:unparam // DB接続処理未実装のため
	f, ok := fetcherRegistry.Get(sourceName)
	if !ok {
		return fmt.Errorf("unknown source: %s", sourceName)
	}

	fmt.Printf("Fetching from %s...\n", sourceName)

	// 汎用的な設定を構築
	config := buildFetchConfig(sourceName)

	posts, err := f.Fetch(config)
	if err != nil {
		return fmt.Errorf("failed to fetch: %w", err)
	}

	fmt.Printf("Found %d posts\n", len(posts))

	if dryRun {
		// ドライランモード：最初の3件を表示
		for i, post := range lo.Slice(posts, 0, 3) {
			fmt.Printf("  [%d] %s (score: %v)\n",
				i+1,
				post.Title,
				post.Engagement["score"])
		}
	} else {
		// TODO: Supabaseに保存
		fmt.Printf("Would save %d posts to Supabase\n", len(posts))
	}

	return nil
}

func buildFetchConfig(sourceName string) fetcher.FetchConfig {
	config := fetcher.FetchConfig{
		Keywords: keywords,
		Limit:    limit,
	}

	// ソース固有の設定
	switch sourceName {
	case "reddit":
		if len(subreddits) == 0 {
			subreddits = []string{"golang", "programming"}
		}
		config.Reddit.Subreddits = subreddits
		// 将来的に追加
		// case "youtube":
		//     config.YouTube.ChannelIDs = channelIDs
	}

	return config
}

func initializeFetchers() {
	fetcherRegistry = fetcher.NewRegistry()

	redditClientID := os.Getenv("REDDIT_CLIENT_ID")
	redditClientSecret := os.Getenv("REDDIT_CLIENT_SECRET")
	redditUsername := os.Getenv("REDDIT_USERNAME")

	fetcherRegistry.Register(reddit.NewRedditFetcher(
		nil,
		redditClientID,
		redditClientSecret,
		redditUsername,
	))
}

func init() {
	rootCmd.AddCommand(fetchCmd)

	fetchCmd.Flags().StringSliceVarP(&sources, "sources", "s", []string{}, "Sources to fetch from (default: all)")
	fetchCmd.Flags().StringSliceVar(&subreddits, "subreddits", []string{}, "Reddit subreddits to fetch")
	fetchCmd.Flags().StringSliceVar(&keywords, "keywords", []string{}, "Keywords to search")
	fetchCmd.Flags().IntVar(&limit, "limit", 25, "Maximum number of items to fetch per source")
	fetchCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Run without saving to database")
}
