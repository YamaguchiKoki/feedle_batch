package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/YamaguchiKoki/feedle_batch/internal/models"
)

type MockFetchedDataRepository struct {
	data          map[string]*models.FetchedData
	lastError     error
	errorOnMethod string
}

func NewMockFetchedDataRepository() *MockFetchedDataRepository {
	return &MockFetchedDataRepository{
		data: make(map[string]*models.FetchedData),
	}
}

func (r *MockFetchedDataRepository) SetError(method string, err error) {
	r.errorOnMethod = method
	r.lastError = err
}

func (r *MockFetchedDataRepository) ClearError() {
	r.errorOnMethod = ""
	r.lastError = nil
}

func (r *MockFetchedDataRepository) GetData() map[string]*models.FetchedData {
	result := make(map[string]*models.FetchedData)
	for k, v := range r.data {
		result[k] = v
	}
	return result
}

func (r *MockFetchedDataRepository) Clear() {
	r.data = make(map[string]*models.FetchedData)
}

func (r *MockFetchedDataRepository) checkError(method string) error {
	if r.errorOnMethod == method || r.errorOnMethod == "all" {
		return r.lastError
	}
	return nil
}

func (r *MockFetchedDataRepository) Create(ctx context.Context, data *models.FetchedData) error {
	if err := r.checkError("Create"); err != nil {
		return err
	}

	if data.FetchedAt.IsZero() {
		data.FetchedAt = time.Now()
	}

	r.data[data.ID] = data
	return nil
}

func (r *MockFetchedDataRepository) CreateBatch(ctx context.Context, data []*models.FetchedData) error {
	if err := r.checkError("CreateBatch"); err != nil {
		return err
	}

	now := time.Now()
	for _, d := range data {
		if d.FetchedAt.IsZero() {
			d.FetchedAt = now
		}
		r.data[d.ID] = d
	}
	return nil
}

func (r *MockFetchedDataRepository) GetByID(ctx context.Context, id string) (*models.FetchedData, error) {
	if err := r.checkError("GetByID"); err != nil {
		return nil, err
	}

	data, exists := r.data[id]
	if !exists {
		return nil, fmt.Errorf("fetched data not found: %s", id)
	}
	return data, nil
}

func (r *MockFetchedDataRepository) GetByConfigID(ctx context.Context, configID string, limit int, offset int) ([]*models.FetchedData, error) {
	if err := r.checkError("GetByConfigID"); err != nil {
		return nil, err
	}

	var results []*models.FetchedData
	for _, data := range r.data {
		if data.ConfigID == configID {
			results = append(results, data)
		}
	}

	if offset > len(results) {
		return []*models.FetchedData{}, nil
	}

	end := len(results)
	if limit > 0 && offset+limit < len(results) {
		end = offset + limit
	}

	return results[offset:end], nil
}

func (r *MockFetchedDataRepository) GetByConfigIDSince(ctx context.Context, configID string, since time.Time, limit int) ([]*models.FetchedData, error) {
	if err := r.checkError("GetByConfigIDSince"); err != nil {
		return nil, err
	}

	var results []*models.FetchedData
	for _, data := range r.data {
		if data.ConfigID == configID && data.FetchedAt.After(since) {
			results = append(results, data)
		}
	}

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

func (r *MockFetchedDataRepository) Update(ctx context.Context, data *models.FetchedData) error {
	if err := r.checkError("Update"); err != nil {
		return err
	}

	if _, exists := r.data[data.ID]; !exists {
		return fmt.Errorf("fetched data not found: %s", data.ID)
	}

	r.data[data.ID] = data
	return nil
}

func (r *MockFetchedDataRepository) Delete(ctx context.Context, id string) error {
	if err := r.checkError("Delete"); err != nil {
		return err
	}

	if _, exists := r.data[id]; !exists {
		return fmt.Errorf("fetched data not found: %s", id)
	}

	delete(r.data, id)
	return nil
}

func (r *MockFetchedDataRepository) ExistsByURL(ctx context.Context, url string) (bool, error) {
	if err := r.checkError("ExistsByURL"); err != nil {
		return false, err
	}

	for _, data := range r.data {
		if data.URL == url {
			return true, nil
		}
	}
	return false, nil
}

func (r *MockFetchedDataRepository) ExistsByRedditID(ctx context.Context, redditID string) (bool, error) {
	if err := r.checkError("ExistsByRedditID"); err != nil {
		return false, err
	}

	for _, data := range r.data {
		if rawData, ok := data.RawData["id"]; ok {
			if redditIDStr, ok := rawData.(string); ok && redditIDStr == redditID {
				return true, nil
			}
		}
	}
	return false, nil
}

func (r *MockFetchedDataRepository) Count(ctx context.Context, configID string) (int64, error) {
	if err := r.checkError("Count"); err != nil {
		return 0, err
	}

	count := int64(0)
	for _, data := range r.data {
		if data.ConfigID == configID {
			count++
		}
	}
	return count, nil
}

type MockConfigRepository struct {
	configs       map[string]*UserFetchConfig
	lastError     error
	errorOnMethod string
}

func NewMockConfigRepository() *MockConfigRepository {
	return &MockConfigRepository{
		configs: make(map[string]*UserFetchConfig),
	}
}

func (r *MockConfigRepository) SetError(method string, err error) {
	r.errorOnMethod = method
	r.lastError = err
}

func (r *MockConfigRepository) checkError(method string) error {
	if r.errorOnMethod == method || r.errorOnMethod == "all" {
		return r.lastError
	}
	return nil
}

func (r *MockConfigRepository) GetUserFetchConfig(ctx context.Context, userID string) (*UserFetchConfig, error) {
	if err := r.checkError("GetUserFetchConfig"); err != nil {
		return nil, err
	}

	for _, config := range r.configs {
		if config.UserID == userID {
			return config, nil
		}
	}
	return nil, fmt.Errorf("config not found for user: %s", userID)
}

func (r *MockConfigRepository) CreateUserFetchConfig(ctx context.Context, config *UserFetchConfig) error {
	if err := r.checkError("CreateUserFetchConfig"); err != nil {
		return err
	}

	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	r.configs[config.ID] = config
	return nil
}

func (r *MockConfigRepository) UpdateUserFetchConfig(ctx context.Context, config *UserFetchConfig) error {
	if err := r.checkError("UpdateUserFetchConfig"); err != nil {
		return err
	}

	if _, exists := r.configs[config.ID]; !exists {
		return fmt.Errorf("config not found: %s", config.ID)
	}

	config.UpdatedAt = time.Now()
	r.configs[config.ID] = config
	return nil
}