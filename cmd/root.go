package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "feedle",
	Short: "Feedle - Multi-source feed aggregator",
	Long: `Feedle is a batch processor that fetches data from multiple sources
including Twitter, YouTube, Instagram, Reddit, and Hacker News.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .env)")
}

func initConfig() {
	// .envファイルを読み込む
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}
