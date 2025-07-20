package models

import "time"

type FetchedData struct {
	ID              string                 `json:"id,omitempty"`
	ConfigID        string                 `json:"config_id"`
	Title           string                 `json:"title"`
	Content         string                 `json:"content"`
	URL             string                 `json:"url"`
	AuthorName      string                 `json:"author_name"`
	AuthorID        string                 `json:"author_id"`
	AuthorAvatarURL string                 `json:"author_avatar_url"`
	PublishedAt     *time.Time             `json:"published_at"`
	Engagement      map[string]interface{} `json:"engagement"`
	Media           []string               `json:"media"`
	Tags            []string               `json:"tags"`
	RawData         map[string]interface{} `json:"raw_data"`
	FetchedAt       time.Time              `json:"fetched_at"`
}
