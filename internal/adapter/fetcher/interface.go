package fetcher

import "github.com/YamaguchiKoki/feedle_batch/internal/domain/model"

type Fetcher interface {
	Name() string

	Fetch(config FetchConfig) ([]*model.FetchedData, error)
}

type FetchConfig struct {
	// Common fields
	Keywords []string
	Limit    int

	// Source-specific fields
	Reddit struct {
		Subreddits []string
	}
	YouTube struct {
		ChannelIDs []string
	}
	HackerNews struct {
		MinScore int
	}
}
