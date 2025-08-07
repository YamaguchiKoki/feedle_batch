package cmd

import (
	"context"
	"log"

	"github.com/YamaguchiKoki/feedle_batch/internal/di"
	"github.com/YamaguchiKoki/feedle_batch/internal/usecase"
	"github.com/samber/do"
	"github.com/spf13/cobra"
)

var (
	dryRun   bool
	injector *do.Injector
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch data from configured sources",
	Long:  `Fetch data from various sources (Reddit, Twitter, etc.) and save to database`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		injector, err = di.NewContainer()
		if err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		uc := do.MustInvoke[*usecase.FetchAndSaveUsecase](injector)

		if err := uc.Execute(ctx); err != nil {
			log.Fatal("Failed to execute fetch:", err)
		}

		log.Println("Fetch completed successfully")
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		if err := injector.Shutdown(); err != nil {
			log.Printf("Warning: Failed to shutdown injector: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)

	fetchCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Run without saving to database")
}
