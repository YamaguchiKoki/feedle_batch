package model

import (
	"time"

	"github.com/google/uuid"
)

type YouTubeFetchConfig struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	UserFetchConfigID uuid.UUID  `json:"user_fetch_config_id" db:"user_fetch_config_id"`
	ChannelID         *string    `json:"channel_id" db:"channel_id"`
	PlaylistID        *string    `json:"playlist_id" db:"playlist_id"`
	Keywords          []string   `json:"keywords" db:"keywords"`
	MaxResults        int        `json:"max_results" db:"max_results"`
	OrderBy           string     `json:"order_by" db:"order_by"`
	PublishedAfter    *time.Time `json:"published_after" db:"published_after"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

func NewYouTubeFetchConfig(userFetchConfigID uuid.UUID, channelID *string, keywords []string) *YouTubeFetchConfig {
	return &YouTubeFetchConfig{
		ID:                uuid.New(),
		UserFetchConfigID: userFetchConfigID,
		ChannelID:         channelID,
		Keywords:          keywords,
		MaxResults:        50,
		OrderBy:           "relevance",
		CreatedAt:         time.Now(),
	}
}
