package repository

import (
	"context"
	"fmt"

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
	var users []struct {
		ID model.UserID `json:"id"`
	}
	_, err := r.client.From("users").Select("id", "", false).ExecuteTo(&users)
	if err != nil {
		return nil, fmt.Errorf("an error occurred during GetActiveUserIDs: %w", err)
	}

	if len(users) == 0 {
		fmt.Println("no record found.")
		return nil, nil
	}

	userIDs := make([]model.UserID, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}
	return userIDs, nil
}
