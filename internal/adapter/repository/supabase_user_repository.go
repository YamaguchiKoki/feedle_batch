package repository

import (
	"context"

	"github.com/YamaguchiKoki/feedle_batch/internal/domain/model"
	"github.com/YamaguchiKoki/feedle_batch/internal/port/output"
	"github.com/supabase-community/supabase-go"
)

type SupabaseUserRepository struct {
	client *supabase.Client
}

func NewSupabaseUserRepository(client *supabase.Client) output.UserRepository {
	return &SupabaseUserRepository{
		client: client,
	}
}

func (r *SupabaseUserRepository) GetActiveUserIDs(ctx context.Context) ([]model.UserID, error) {
	// TODO: 実装
	return nil, nil
}
