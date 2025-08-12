package model

import (
	"encoding/json"
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

// UnmarshalJSON custom unmarshaler to handle Supabase timestamp format
func (u *UserFetchConfig) UnmarshalJSON(data []byte) error {
	// Temporary struct with string timestamps
	aux := &struct {
		ID           uuid.UUID `json:"id"`
		UserID       uuid.UUID `json:"user_id"`
		Name         string    `json:"name"`
		DataSourceID string    `json:"data_source_id"`
		IsActive     bool      `json:"is_active"`
		CreatedAt    string    `json:"created_at"`
		UpdatedAt    string    `json:"updated_at"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	u.ID = aux.ID
	u.UserID = aux.UserID
	u.Name = aux.Name
	u.DataSourceID = aux.DataSourceID
	u.IsActive = aux.IsActive

	// Parse timestamps without timezone (take first 19 chars)
	if len(aux.CreatedAt) >= 19 {
		t, err := time.Parse("2006-01-02T15:04:05", aux.CreatedAt[:19])
		if err != nil {
			return err
		}
		u.CreatedAt = t
	}

	if len(aux.UpdatedAt) >= 19 {
		t, err := time.Parse("2006-01-02T15:04:05", aux.UpdatedAt[:19])
		if err != nil {
			return err
		}
		u.UpdatedAt = t
	}

	return nil
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
