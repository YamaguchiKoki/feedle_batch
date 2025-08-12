package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type FetchConfigDetail interface {
	GetUserFetchConfigID() uuid.UUID
	GetDataSourceID() string
}

type RedditFetchConfigDetail struct {
	ID                uuid.UUID `json:"id" db:"id"`
	UserFetchConfigID uuid.UUID `json:"user_fetch_config_id" db:"user_fetch_config_id"`
	Subreddit         string    `json:"subreddit" db:"subreddit"`
	SortBy            string    `json:"sort_by" db:"sort_by"`
	TimeFilter        string    `json:"time_filter" db:"time_filter"`
	LimitCount        int       `json:"limit_count" db:"limit_count"`
	Keywords          []string  `json:"keywords" db:"keywords"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

// UnmarshalJSON custom unmarshaler to handle Supabase timestamp format
func (r *RedditFetchConfigDetail) UnmarshalJSON(data []byte) error {
	// Temporary struct with string timestamp
	aux := &struct {
		ID                uuid.UUID `json:"id"`
		UserFetchConfigID uuid.UUID `json:"user_fetch_config_id"`
		Subreddit         string    `json:"subreddit"`
		SortBy            string    `json:"sort_by"`
		TimeFilter        string    `json:"time_filter"`
		LimitCount        int       `json:"limit_count"`
		Keywords          []string  `json:"keywords"`
		CreatedAt         string    `json:"created_at"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	r.ID = aux.ID
	r.UserFetchConfigID = aux.UserFetchConfigID
	r.Subreddit = aux.Subreddit
	r.SortBy = aux.SortBy
	r.TimeFilter = aux.TimeFilter
	r.LimitCount = aux.LimitCount
	r.Keywords = aux.Keywords

	// Parse timestamp without timezone (take first 19 chars)
	if len(aux.CreatedAt) >= 19 {
		t, err := time.Parse("2006-01-02T15:04:05", aux.CreatedAt[:19])
		if err != nil {
			return err
		}
		r.CreatedAt = t
	}

	return nil
}

func (r RedditFetchConfigDetail) GetUserFetchConfigID() uuid.UUID {
	return r.UserFetchConfigID
}

func (r RedditFetchConfigDetail) GetDataSourceID() string {
	return "reddit"
}
