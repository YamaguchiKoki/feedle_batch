package model

import (
	"time"

	"github.com/google/uuid"
)

type UserFetchConfig struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	Name         string    `json:"name" db:"name"`
	DataSourceID string    `json:"data_source_id" db:"data_source_id"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

func NewUserFetchConfig(userID uuid.UUID, name string, dataSourceID string) *UserFetchConfig {
	now := time.Now()
	return &UserFetchConfig{
		ID:           uuid.New(),
		UserID:       userID,
		Name:         name,
		DataSourceID: dataSourceID,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
