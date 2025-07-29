package repository

import (
	"context"
	"testing"
	"time"

	"github.com/YamaguchiKoki/feedle_batch/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockFetchedDataRepository(t *testing.T) {
	repo := NewMockFetchedDataRepository()

	t.Run("Create", func(t *testing.T) {
		data := &models.FetchedData{
			ID:       "test-1",
			ConfigID: "config-1",
			Title:    "Test Post",
			Content:  "Test content",
			URL:      "https://example.com/test-1",
		}

		err := repo.Create(context.Background(), data)
		require.NoError(t, err)

		stored := repo.GetData()
		assert.Len(t, stored, 1)
		assert.Equal(t, data.Title, stored["test-1"].Title)
		assert.False(t, stored["test-1"].FetchedAt.IsZero())
	})

	t.Run("CreateBatch", func(t *testing.T) {
		repo.Clear()

		data := []*models.FetchedData{
			{ID: "batch-1", ConfigID: "config-1", Title: "Batch 1"},
			{ID: "batch-2", ConfigID: "config-1", Title: "Batch 2"},
		}

		err := repo.CreateBatch(context.Background(), data)
		require.NoError(t, err)

		stored := repo.GetData()
		assert.Len(t, stored, 2)
	})

	t.Run("GetByID", func(t *testing.T) {
		repo.Clear()

		original := &models.FetchedData{
			ID:       "get-test",
			ConfigID: "config-1",
			Title:    "Get Test",
		}

		err := repo.Create(context.Background(), original)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(context.Background(), "get-test")
		require.NoError(t, err)
		assert.Equal(t, original.Title, retrieved.Title)

		_, err = repo.GetByID(context.Background(), "non-existent")
		assert.Error(t, err)
	})

	t.Run("GetByConfigID", func(t *testing.T) {
		repo.Clear()

		data := []*models.FetchedData{
			{ID: "config1-1", ConfigID: "config-1", Title: "Config 1 Post 1"},
			{ID: "config1-2", ConfigID: "config-1", Title: "Config 1 Post 2"},
			{ID: "config2-1", ConfigID: "config-2", Title: "Config 2 Post 1"},
		}

		for _, d := range data {
			err := repo.Create(context.Background(), d)
			require.NoError(t, err)
		}

		results, err := repo.GetByConfigID(context.Background(), "config-1", 10, 0)
		require.NoError(t, err)
		assert.Len(t, results, 2)

		results, err = repo.GetByConfigID(context.Background(), "config-1", 1, 0)
		require.NoError(t, err)
		assert.Len(t, results, 1)

		results, err = repo.GetByConfigID(context.Background(), "config-1", 1, 1)
		require.NoError(t, err)
		assert.Len(t, results, 1)
	})

	t.Run("GetByConfigIDSince", func(t *testing.T) {
		repo.Clear()

		now := time.Now()
		past := now.Add(-time.Hour)
		future := now.Add(time.Hour)

		data := []*models.FetchedData{
			{ID: "past", ConfigID: "config-1", Title: "Past", FetchedAt: past},
			{ID: "present", ConfigID: "config-1", Title: "Present", FetchedAt: now},
			{ID: "future", ConfigID: "config-1", Title: "Future", FetchedAt: future},
		}

		for _, d := range data {
			err := repo.Create(context.Background(), d)
			require.NoError(t, err)
		}

		results, err := repo.GetByConfigIDSince(context.Background(), "config-1", now.Add(-30*time.Minute), 10)
		require.NoError(t, err)
		assert.Len(t, results, 2) // present and future
	})

	t.Run("Update", func(t *testing.T) {
		repo.Clear()

		original := &models.FetchedData{
			ID:       "update-test",
			ConfigID: "config-1",
			Title:    "Original Title",
		}

		err := repo.Create(context.Background(), original)
		require.NoError(t, err)

		updated := &models.FetchedData{
			ID:       "update-test",
			ConfigID: "config-1",
			Title:    "Updated Title",
		}

		err = repo.Update(context.Background(), updated)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(context.Background(), "update-test")
		require.NoError(t, err)
		assert.Equal(t, "Updated Title", retrieved.Title)

		nonExistent := &models.FetchedData{ID: "non-existent"}
		err = repo.Update(context.Background(), nonExistent)
		assert.Error(t, err)
	})

	t.Run("Delete", func(t *testing.T) {
		repo.Clear()

		data := &models.FetchedData{
			ID:       "delete-test",
			ConfigID: "config-1",
			Title:    "Delete Test",
		}

		err := repo.Create(context.Background(), data)
		require.NoError(t, err)

		err = repo.Delete(context.Background(), "delete-test")
		require.NoError(t, err)

		_, err = repo.GetByID(context.Background(), "delete-test")
		assert.Error(t, err)

		err = repo.Delete(context.Background(), "non-existent")
		assert.Error(t, err)
	})

	t.Run("ExistsByURL", func(t *testing.T) {
		repo.Clear()

		data := &models.FetchedData{
			ID:       "url-test",
			ConfigID: "config-1",
			Title:    "URL Test",
			URL:      "https://example.com/unique-url",
		}

		err := repo.Create(context.Background(), data)
		require.NoError(t, err)

		exists, err := repo.ExistsByURL(context.Background(), "https://example.com/unique-url")
		require.NoError(t, err)
		assert.True(t, exists)

		exists, err = repo.ExistsByURL(context.Background(), "https://example.com/non-existent")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("ExistsByRedditID", func(t *testing.T) {
		repo.Clear()

		data := &models.FetchedData{
			ID:       "reddit-test",
			ConfigID: "config-1",
			Title:    "Reddit Test",
			RawData: map[string]interface{}{
				"id": "reddit123",
			},
		}

		err := repo.Create(context.Background(), data)
		require.NoError(t, err)

		exists, err := repo.ExistsByRedditID(context.Background(), "reddit123")
		require.NoError(t, err)
		assert.True(t, exists)

		exists, err = repo.ExistsByRedditID(context.Background(), "reddit456")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Count", func(t *testing.T) {
		repo.Clear()

		data := []*models.FetchedData{
			{ID: "count1", ConfigID: "config-1", Title: "Count 1"},
			{ID: "count2", ConfigID: "config-1", Title: "Count 2"},
			{ID: "count3", ConfigID: "config-2", Title: "Count 3"},
		}

		for _, d := range data {
			err := repo.Create(context.Background(), d)
			require.NoError(t, err)
		}

		count, err := repo.Count(context.Background(), "config-1")
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)

		count, err = repo.Count(context.Background(), "config-2")
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		count, err = repo.Count(context.Background(), "non-existent")
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		repo.Clear()

		testErr := assert.AnError
		repo.SetError("Create", testErr)

		data := &models.FetchedData{ID: "error-test"}
		err := repo.Create(context.Background(), data)
		assert.Equal(t, testErr, err)

		repo.ClearError()
		err = repo.Create(context.Background(), data)
		assert.NoError(t, err)
	})
}