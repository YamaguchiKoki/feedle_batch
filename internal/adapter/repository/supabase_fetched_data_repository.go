package repository

import (
	"context"
	"fmt"

	"github.com/YamaguchiKoki/feedle_batch/internal/domain/model"
	"github.com/YamaguchiKoki/feedle_batch/internal/port/output"
	"github.com/google/uuid"
	"github.com/supabase-community/supabase-go"
)

type SupabaseFetchedDataRepository struct {
	client *supabase.Client
}

func NewSupabaseFetchedDataRepository(client *supabase.Client) output.FetchedDataRepository {
	return &SupabaseFetchedDataRepository{
		client: client,
	}
}

func (r *SupabaseFetchedDataRepository) Create(ctx context.Context, data *model.FetchedData) error {
	// Supabaseに保存するための構造体（時刻フィールドを文字列に変換）
	type fetchedDataInsert struct {
		ID           uuid.UUID              `json:"id"`
		ConfigID     uuid.UUID              `json:"config_id"`
		Source       string                 `json:"source"`
		Title        string                 `json:"title"`
		Content      string                 `json:"content,omitempty"`
		URL          string                 `json:"url,omitempty"`
		AuthorName   string                 `json:"author_name,omitempty"`
		SourceItemID string                 `json:"source_item_id,omitempty"`
		PublishedAt  *string                `json:"published_at,omitempty"`
		Tags         []string               `json:"tags"`
		MediaURLs    []string               `json:"media_urls"`
		Metadata     map[string]interface{} `json:"metadata"`
		FetchedAt    string                 `json:"fetched_at"`
		CreatedAt    string                 `json:"created_at"`
	}

	// 時刻をISO 8601形式の文字列に変換
	insertData := fetchedDataInsert{
		ID:           data.ID,
		ConfigID:     data.ConfigID,
		Source:       data.Source,
		Title:        data.Title,
		Content:      data.Content,
		URL:          data.URL,
		AuthorName:   data.AuthorName,
		SourceItemID: data.SourceItemID,
		Tags:         data.Tags,
		MediaURLs:    data.MediaURLs,
		Metadata:     data.Metadata,
		FetchedAt:    data.FetchedAt.Format("2006-01-02T15:04:05"),
		CreatedAt:    data.CreatedAt.Format("2006-01-02T15:04:05"),
	}

	// PublishedAtがnilでない場合のみ変換
	if data.PublishedAt != nil {
		publishedAtStr := data.PublishedAt.Format("2006-01-02T15:04:05")
		insertData.PublishedAt = &publishedAtStr
	}

	// Supabaseにデータを挿入
	_, err := r.client.From("fetched_data").Insert(insertData, false, "", "", "").ExecuteTo(nil)
	if err != nil {
		return fmt.Errorf("failed to insert fetched data: %w", err)
	}

	return nil
}
