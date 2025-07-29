package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/YamaguchiKoki/feedle_batch/internal/models"
	"github.com/supabase-community/supabase-go"
)

type SupabaseFetchedDataRepository struct {
	client *supabase.Client
}

func NewSupabaseFetchedDataRepository(client *supabase.Client) FetchedDataRepository {
	return &SupabaseFetchedDataRepository{
		client: client,
	}
}

func (r *SupabaseFetchedDataRepository) Create(ctx context.Context, data *models.FetchedData) error {
	if data.FetchedAt.IsZero() {
		data.FetchedAt = time.Now()
	}

	_, _, err := r.client.From("fetched_data").Insert(data, false, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("failed to create fetched data: %w", err)
	}

	return nil
}

func (r *SupabaseFetchedDataRepository) CreateBatch(ctx context.Context, data []*models.FetchedData) error {
	if len(data) == 0 {
		return nil
	}

	now := time.Now()
	for _, d := range data {
		if d.FetchedAt.IsZero() {
			d.FetchedAt = now
		}
	}

	_, _, err := r.client.From("fetched_data").Insert(data, false, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("failed to create batch fetched data: %w", err)
	}

	return nil
}

func (r *SupabaseFetchedDataRepository) GetByID(ctx context.Context, id string) (*models.FetchedData, error) {
	data, _, err := r.client.From("fetched_data").
		Select("*", "", false).
		Eq("id", id).
		Single().
		Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get fetched data by ID: %w", err)
	}

	var result models.FetchedData
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal fetched data: %w", err)
	}

	return &result, nil
}

func (r *SupabaseFetchedDataRepository) GetByConfigID(ctx context.Context, configID string, limit int, offset int) ([]*models.FetchedData, error) {
	query := r.client.From("fetched_data").
		Select("*", "", false).
		Eq("config_id", configID).
		Order("fetched_at", nil)

	if limit > 0 {
		query = query.Limit(limit, "")
	}
	if offset > 0 {
		query = query.Range(offset, offset+limit-1, "")
	}

	data, _, err := query.Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get fetched data by config ID: %w", err)
	}

	var results []*models.FetchedData
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("failed to unmarshal fetched data: %w", err)
	}

	return results, nil
}

func (r *SupabaseFetchedDataRepository) GetByConfigIDSince(ctx context.Context, configID string, since time.Time, limit int) ([]*models.FetchedData, error) {
	query := r.client.From("fetched_data").
		Select("*", "", false).
		Eq("config_id", configID).
		Gte("fetched_at", since.Format(time.RFC3339)).
		Order("fetched_at", nil)

	if limit > 0 {
		query = query.Limit(limit, "")
	}

	data, _, err := query.Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get fetched data by config ID since: %w", err)
	}

	var results []*models.FetchedData
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("failed to unmarshal fetched data: %w", err)
	}

	return results, nil
}

func (r *SupabaseFetchedDataRepository) Update(ctx context.Context, data *models.FetchedData) error {
	_, _, err := r.client.From("fetched_data").
		Update(data, "", "").
		Eq("id", data.ID).
		Execute()
	if err != nil {
		return fmt.Errorf("failed to update fetched data: %w", err)
	}

	return nil
}

func (r *SupabaseFetchedDataRepository) Delete(ctx context.Context, id string) error {
	_, _, err := r.client.From("fetched_data").
		Delete("", "").
		Eq("id", id).
		Execute()
	if err != nil {
		return fmt.Errorf("failed to delete fetched data: %w", err)
	}

	return nil
}

func (r *SupabaseFetchedDataRepository) ExistsByURL(ctx context.Context, url string) (bool, error) {
	data, _, err := r.client.From("fetched_data").
		Select("id", "", false).
		Eq("url", url).
		Limit(1, "").
		Execute()
	if err != nil {
		return false, fmt.Errorf("failed to check existence by URL: %w", err)
	}

	var results []map[string]interface{}
	if err := json.Unmarshal(data, &results); err != nil {
		return false, fmt.Errorf("failed to unmarshal existence check result: %w", err)
	}

	return len(results) > 0, nil
}

func (r *SupabaseFetchedDataRepository) ExistsByRedditID(ctx context.Context, redditID string) (bool, error) {
	data, _, err := r.client.From("fetched_data").
		Select("id", "", false).
		Eq("raw_data->>id", redditID).
		Limit(1, "").
		Execute()
	if err != nil {
		return false, fmt.Errorf("failed to check existence by Reddit ID: %w", err)
	}

	var results []map[string]interface{}
	if err := json.Unmarshal(data, &results); err != nil {
		return false, fmt.Errorf("failed to unmarshal existence check result: %w", err)
	}

	return len(results) > 0, nil
}

func (r *SupabaseFetchedDataRepository) Count(ctx context.Context, configID string) (int64, error) {
	data, _, err := r.client.From("fetched_data").
		Select("id", "", false).
		Eq("config_id", configID).
		Execute()
	if err != nil {
		return 0, fmt.Errorf("failed to count fetched data: %w", err)
	}

	var results []map[string]interface{}
	if err := json.Unmarshal(data, &results); err != nil {
		return 0, fmt.Errorf("failed to unmarshal count result: %w", err)
	}

	return int64(len(results)), nil
}