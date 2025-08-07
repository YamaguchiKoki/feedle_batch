package repository

import (
	"context"

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

// GetByUserID は指定されたユーザーIDの全設定を取得します
func (r *SupabaseFetchConfigRepository) GetByUserID(ctx context.Context, userID model.UserID) ([]*model.UserFetchConfig, error) {
	// TODO: 実装
	return nil, nil
}
