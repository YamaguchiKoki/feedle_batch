package repository

import (
	"context"

	"github.com/YamaguchiKoki/feedle_batch/internal/domain/model"
	"github.com/YamaguchiKoki/feedle_batch/internal/port/output"
	"github.com/supabase-community/supabase-go"
)

type SupabaseDataSourceRepository struct {
	client *supabase.Client
}

func NewSupabaseDataSourceRepository(client *supabase.Client) output.DataSourceRepository {
	return &SupabaseDataSourceRepository{
		client: client,
	}
}

func (r *SupabaseDataSourceRepository) GetByID(ctx context.Context, id string) (*model.DataSource, error) {
	var dataSource model.DataSource
	_, err := r.client.From("data_sources").Select("*", "", false).Eq("id", id).Single().ExecuteTo(&dataSource)
	if err != nil {
		return nil, err
	}
	return &dataSource, nil
}

func (r *SupabaseDataSourceRepository) GetAll(ctx context.Context) ([]*model.DataSource, error) {
	var dataSources []*model.DataSource
	_, err := r.client.From("data_sources").Select("*", "", false).ExecuteTo(&dataSources)
	if err != nil {
		return nil, err
	}
	return dataSources, nil
}

func (r *SupabaseDataSourceRepository) GetActive(ctx context.Context) ([]*model.DataSource, error) {
	var dataSources []*model.DataSource
	_, err := r.client.From("data_sources").Select("*", "", false).Eq("is_active", "true").ExecuteTo(&dataSources)
	if err != nil {
		return nil, err
	}
	return dataSources, nil
}
