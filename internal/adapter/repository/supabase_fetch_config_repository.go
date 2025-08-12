package repository

import (
	"context"
	"fmt"

	"github.com/YamaguchiKoki/feedle_batch/internal/domain/model"
	"github.com/YamaguchiKoki/feedle_batch/internal/port/output"
	"github.com/supabase-community/supabase-go"
)

type SupabaseFetchConfigRepository struct {
	client *supabase.Client
}

func NewSupabaseFetchConfigRepository(client *supabase.Client) output.FetchConfigRepository {
	return &SupabaseFetchConfigRepository{
		client: client,
	}
}

func (r *SupabaseFetchConfigRepository) GetByUserID(ctx context.Context, userID model.UserID) ([]model.UserFetchConfig, error) {
	var fetchConfigs []model.UserFetchConfig
	_, err := r.client.From("user_fetch_configs").Select("*", "", false).Eq("is_active", "true").ExecuteTo(&fetchConfigs)
	if err != nil {
		return nil, fmt.Errorf("an error occurred during GetByUserID(user_fetch_config): %w", err)
	}
	if len(fetchConfigs) == 0 {
		fmt.Printf("no configs found")
		return nil, nil
	}
	return fetchConfigs, nil
}
