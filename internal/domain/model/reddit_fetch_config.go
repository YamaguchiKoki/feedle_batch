package model

import (
	"time"

	"github.com/google/uuid"
)

type RedditFetchConfig struct {
	ID                uuid.UUID `json:"id" db:"id"`
	UserFetchConfigID uuid.UUID `json:"user_fetch_config_id" db:"user_fetch_config_id"`
	Subreddit         *string   `json:"subreddit" db:"subreddit"`
	SortBy            string    `json:"sort_by" db:"sort_by"`
	TimeFilter        string    `json:"time_filter" db:"time_filter"`
	LimitCount        int       `json:"limit_count" db:"limit_count"`
	Keywords          []string  `json:"keywords" db:"keywords"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

func NewRedditFetchConfig(userFetchConfigID uuid.UUID, subreddit *string, keywords []string) *RedditFetchConfig {
	return &RedditFetchConfig{
		ID:                uuid.New(),
		UserFetchConfigID: userFetchConfigID,
		Subreddit:         subreddit,
		SortBy:            "hot",
		TimeFilter:        "day",
		LimitCount:        25,
		Keywords:          keywords,
		CreatedAt:         time.Now(),
	}
}
