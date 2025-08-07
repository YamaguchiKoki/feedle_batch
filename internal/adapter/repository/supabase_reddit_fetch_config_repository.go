package repository

import (
	"context"

	"github.com/YamaguchiKoki/feedle_batch/internal/domain/model"
	"github.com/YamaguchiKoki/feedle_batch/internal/port/output"
	"github.com/google/uuid"
	"github.com/supabase-community/supabase-go"
)

type SupabaseRedditFetchConfigRepository struct {
	client *supabase.Client
}

func NewSupabaseRedditFetchConfigRepository(client *supabase.Client) output.RedditFetchConfigRepository {
	return &SupabaseRedditFetchConfigRepository{
		client: client,
	}
}

func (r *SupabaseRedditFetchConfigRepository) GetByUserFetchConfigID(ctx context.Context, userFetchConfigID uuid.UUID) (*model.RedditFetchConfigDetail, error) {
	var config model.RedditFetchConfigDetail
	_, err := r.client.From("reddit_fetch_configs").Select("*", "", false).Eq("user_fetch_config_id", userFetchConfigID.String()).Single().ExecuteTo(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
