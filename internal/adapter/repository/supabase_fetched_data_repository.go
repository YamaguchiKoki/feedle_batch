package repository

import (
	"context"

	"github.com/YamaguchiKoki/feedle_batch/internal/domain/model"
	"github.com/YamaguchiKoki/feedle_batch/internal/port/output"
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
	// TODO: 実装
	return nil
}
