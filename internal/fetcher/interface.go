package fetcher

import "github.com/YamaguchiKoki/feedle_batch/internal/models"

type Fetcher interface {
	Name() string

	Fetch(config FetchConfig) ([]*models.FetchedData, error)
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

type Registry struct {
	fetchers map[string]Fetcher
}

func NewRegistry() *Registry {
	return &Registry{
		fetchers: make(map[string]Fetcher),
	}
}

func (r *Registry) Register(fetcher Fetcher) {
	r.fetchers[fetcher.Name()] = fetcher
}

func (r *Registry) Get(name string) (Fetcher, bool) {
	f, ok := r.fetchers[name]
	return f, ok
}

func (r *Registry) GetAll() []Fetcher {
	fetchers := make([]Fetcher, 0, len(r.fetchers))
	for _, f := range r.fetchers {
		fetchers = append(fetchers, f)
	}
	return fetchers
}
