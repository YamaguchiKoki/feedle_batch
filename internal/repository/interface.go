package repository

import (
	"context"
	"time"

	"github.com/YamaguchiKoki/feedle_batch/internal/models"
)

type FetchedDataRepository interface {
	Create(ctx context.Context, data *models.FetchedData) error
	CreateBatch(ctx context.Context, data []*models.FetchedData) error
	GetByID(ctx context.Context, id string) (*models.FetchedData, error)
	GetByConfigID(ctx context.Context, configID string, limit int, offset int) ([]*models.FetchedData, error)
	GetByConfigIDSince(ctx context.Context, configID string, since time.Time, limit int) ([]*models.FetchedData, error)
	Update(ctx context.Context, data *models.FetchedData) error
	Delete(ctx context.Context, id string) error
	ExistsByURL(ctx context.Context, url string) (bool, error)
	ExistsByRedditID(ctx context.Context, redditID string) (bool, error)
	Count(ctx context.Context, configID string) (int64, error)
}

type ConfigRepository interface {
	GetUserFetchConfig(ctx context.Context, userID string) (*UserFetchConfig, error)
	CreateUserFetchConfig(ctx context.Context, config *UserFetchConfig) error
	UpdateUserFetchConfig(ctx context.Context, config *UserFetchConfig) error
}

type UserFetchConfig struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	IsActive    bool                   `json:"is_active"`
	Config      map[string]interface{} `json:"config"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}