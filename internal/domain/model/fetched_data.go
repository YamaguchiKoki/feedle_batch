package model

import (
	"time"

	"github.com/google/uuid"
)

type FetchedData struct {
	ID           uuid.UUID              `json:"id"`
	ConfigID     uuid.UUID              `json:"config_id"`
	Source       string                 `json:"source"`
	Title        string                 `json:"title"`
	Content      string                 `json:"content,omitempty"`
	URL          string                 `json:"url,omitempty"`
	AuthorName   string                 `json:"author_name,omitempty"`
	SourceItemID string                 `json:"source_item_id,omitempty"`
	PublishedAt  *time.Time             `json:"published_at,omitempty"`
	Tags         []string               `json:"tags"`
	MediaURLs    []string               `json:"media_urls"`
	Metadata     map[string]interface{} `json:"metadata"`
	FetchedAt    time.Time              `json:"fetched_at"`
	CreatedAt    time.Time              `json:"created_at"`
}
