package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/YamaguchiKoki/feedle_batch/internal/models"
	"github.com/YamaguchiKoki/feedle_batch/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataService_SaveFetchedData(t *testing.T) {
	t.Run("SaveNewData", func(t *testing.T) {
		repo := repository.NewMockFetchedDataRepository()
		service := NewDataService(repo)

		data := []*models.FetchedData{
			{
				ID:       "test-1",
				Title:    "Test Post 1",
				URL:      "https://example.com/test-1",
				RawData:  map[string]interface{}{"id": "reddit1"},
			},
			{
				ID:       "test-2",
				Title:    "Test Post 2",
				URL:      "https://example.com/test-2",
				RawData:  map[string]interface{}{"id": "reddit2"},
			},
		}

		opts := SaveOptions{
			ConfigID:            "test-config",
			SkipDuplicatesByURL: true,
			BatchSize:           10,
		}

		result, err := service.SaveFetchedData(context.Background(), data, opts)

		require.NoError(t, err)
		assert.Equal(t, 2, result.Total)
		assert.Equal(t, 2, result.Saved)
		assert.Equal(t, 0, result.Duplicates)
		assert.Equal(t, 0, result.Skipped)
		assert.Empty(t, result.Errors)
		assert.True(t, result.Success())

		stored := repo.GetData()
		assert.Len(t, stored, 2)
		assert.Equal(t, "test-config", stored["test-1"].ConfigID)
	})

	t.Run("SkipDuplicatesByURL", func(t *testing.T) {
		repo := repository.NewMockFetchedDataRepository()
		service := NewDataService(repo)

		existing := &models.FetchedData{
			ID:       "existing",
			Title:    "Existing Post",
			URL:      "https://example.com/duplicate",
			RawData:  map[string]interface{}{"id": "reddit_existing"},
		}
		err := repo.Create(context.Background(), existing)
		require.NoError(t, err)

		data := []*models.FetchedData{
			{
				ID:       "new-1",
				Title:    "New Post",
				URL:      "https://example.com/new",
				RawData:  map[string]interface{}{"id": "reddit_new"},
			},
			{
				ID:       "duplicate",
				Title:    "Duplicate Post",
				URL:      "https://example.com/duplicate",
				RawData:  map[string]interface{}{"id": "reddit_dup"},
			},
		}

		opts := SaveOptions{
			ConfigID:            "test-config",
			SkipDuplicatesByURL: true,
			BatchSize:           10,
		}

		result, err := service.SaveFetchedData(context.Background(), data, opts)

		require.NoError(t, err)
		assert.Equal(t, 2, result.Total)
		assert.Equal(t, 1, result.Saved)
		assert.Equal(t, 1, result.Duplicates)
		assert.Equal(t, 0, result.Skipped)

		stored := repo.GetData()
		assert.Len(t, stored, 2) // existing + new-1
		assert.Contains(t, stored, "existing")
		assert.Contains(t, stored, "new-1")
		assert.NotContains(t, stored, "duplicate")
	})

	t.Run("SkipDuplicatesByRedditID", func(t *testing.T) {
		repo := repository.NewMockFetchedDataRepository()
		service := NewDataService(repo)

		existing := &models.FetchedData{
			ID:       "existing",
			Title:    "Existing Post",
			URL:      "https://example.com/existing",
			RawData:  map[string]interface{}{"id": "reddit_duplicate"},
		}
		err := repo.Create(context.Background(), existing)
		require.NoError(t, err)

		data := []*models.FetchedData{
			{
				ID:       "new",
				Title:    "New Post",
				URL:      "https://example.com/new",
				RawData:  map[string]interface{}{"id": "reddit_new"},
			},
			{
				ID:       "duplicate",
				Title:    "Duplicate Post",
				URL:      "https://example.com/different-url",
				RawData:  map[string]interface{}{"id": "reddit_duplicate"},
			},
		}

		opts := SaveOptions{
			ConfigID:            "test-config",
			SkipDuplicatesByURL: false,
			BatchSize:           10,
		}

		result, err := service.SaveFetchedData(context.Background(), data, opts)

		require.NoError(t, err)
		assert.Equal(t, 2, result.Total)
		assert.Equal(t, 1, result.Saved)
		assert.Equal(t, 1, result.Duplicates)

		stored := repo.GetData()
		assert.Len(t, stored, 2)
		assert.Contains(t, stored, "existing")
		assert.Contains(t, stored, "new")
		assert.NotContains(t, stored, "duplicate")
	})

	t.Run("EmptyData", func(t *testing.T) {
		repo := repository.NewMockFetchedDataRepository()
		service := NewDataService(repo)

		opts := SaveOptions{ConfigID: "test-config"}
		result, err := service.SaveFetchedData(context.Background(), []*models.FetchedData{}, opts)

		require.NoError(t, err)
		assert.Equal(t, 0, result.Total)
		assert.Equal(t, 0, result.Saved)
	})

	t.Run("BatchProcessing", func(t *testing.T) {
		repo := repository.NewMockFetchedDataRepository()
		service := NewDataService(repo)

		data := make([]*models.FetchedData, 5)
		for i := 0; i < 5; i++ {
			data[i] = &models.FetchedData{
				ID:      fmt.Sprintf("batch-%d", i),
				Title:   fmt.Sprintf("Batch Post %d", i),
				URL:     fmt.Sprintf("https://example.com/batch-%d", i),
				RawData: map[string]interface{}{"id": fmt.Sprintf("reddit_batch_%d", i)},
			}
		}

		opts := SaveOptions{
			ConfigID:  "test-config",
			BatchSize: 2,
		}

		result, err := service.SaveFetchedData(context.Background(), data, opts)

		require.NoError(t, err)
		assert.Equal(t, 5, result.Total)
		assert.Equal(t, 5, result.Saved)

		stored := repo.GetData()
		assert.Len(t, stored, 5)
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		repo := repository.NewMockFetchedDataRepository()
		service := NewDataService(repo)

		repo.SetError("CreateBatch", assert.AnError)

		data := []*models.FetchedData{
			{
				ID:      "error-test",
				Title:   "Error Test",
				URL:     "https://example.com/error",
				RawData: map[string]interface{}{"id": "reddit_error"},
			},
		}

		opts := SaveOptions{ConfigID: "test-config"}
		result, err := service.SaveFetchedData(context.Background(), data, opts)

		require.NoError(t, err)
		assert.Equal(t, 1, result.Total)
		assert.Equal(t, 0, result.Saved)
		assert.Len(t, result.Errors, 1)
		assert.False(t, result.Success())
	})
}

func TestDataService_GetRecentData(t *testing.T) {
	repo := repository.NewMockFetchedDataRepository()
	service := NewDataService(repo)

	data := []*models.FetchedData{
		{ID: "recent-1", ConfigID: "config-1", Title: "Recent 1"},
		{ID: "recent-2", ConfigID: "config-1", Title: "Recent 2"},
		{ID: "recent-3", ConfigID: "config-2", Title: "Recent 3"},
	}

	for _, d := range data {
		err := repo.Create(context.Background(), d)
		require.NoError(t, err)
	}

	results, err := service.GetRecentData(context.Background(), "config-1", 10)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestDataService_GetDataSince(t *testing.T) {
	repo := repository.NewMockFetchedDataRepository()
	service := NewDataService(repo)

	now := time.Now()
	past := now.Add(-time.Hour)

	data := []*models.FetchedData{
		{ID: "since-1", ConfigID: "config-1", Title: "Since 1", FetchedAt: past},
		{ID: "since-2", ConfigID: "config-1", Title: "Since 2", FetchedAt: now},
	}

	for _, d := range data {
		err := repo.Create(context.Background(), d)
		require.NoError(t, err)
	}

	results, err := service.GetDataSince(context.Background(), "config-1", past.Add(30*time.Minute), 10)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Since 2", results[0].Title)
}

func TestDataService_GetStats(t *testing.T) {
	repo := repository.NewMockFetchedDataRepository()
	service := NewDataService(repo)

	now := time.Now()
	data := []*models.FetchedData{
		{ID: "stats-1", ConfigID: "config-1", Title: "Stats 1", FetchedAt: now.Add(-time.Hour)},
		{ID: "stats-2", ConfigID: "config-1", Title: "Stats 2", FetchedAt: now},
	}

	for _, d := range data {
		err := repo.Create(context.Background(), d)
		require.NoError(t, err)
	}

	stats, err := service.GetStats(context.Background(), "config-1")
	require.NoError(t, err)
	assert.Equal(t, int64(2), stats.TotalCount)
	assert.NotNil(t, stats.LastFetchedAt)
}

func TestSaveResult(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		result := &SaveResult{
			Total:   5,
			Saved:   5,
			Errors:  []string{},
		}
		assert.True(t, result.Success())
		assert.Equal(t, "Total: 5, Saved: 5, Duplicates: 0, Skipped: 0", result.Summary())
	})

	t.Run("WithErrors", func(t *testing.T) {
		result := &SaveResult{
			Total:      5,
			Saved:      3,
			Duplicates: 1,
			Skipped:    0,
			Errors:     []string{"error 1", "error 2"},
		}
		assert.False(t, result.Success())
		assert.Equal(t, "Total: 5, Saved: 3, Duplicates: 1, Skipped: 0, Errors: 2", result.Summary())
	})
}