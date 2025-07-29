package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YamaguchiKoki/feedle_batch/internal/models"
	"github.com/YamaguchiKoki/feedle_batch/internal/repository"
)

type DataService struct {
	fetchedDataRepo repository.FetchedDataRepository
}

func NewDataService(fetchedDataRepo repository.FetchedDataRepository) *DataService {
	return &DataService{
		fetchedDataRepo: fetchedDataRepo,
	}
}

type SaveOptions struct {
	ConfigID            string
	SkipDuplicatesByURL bool
	BatchSize           int
}

func (s *DataService) SaveFetchedData(ctx context.Context, data []*models.FetchedData, opts SaveOptions) (*SaveResult, error) {
	if len(data) == 0 {
		return &SaveResult{}, nil
	}

	if opts.BatchSize <= 0 {
		opts.BatchSize = 100
	}

	result := &SaveResult{
		Total:      len(data),
		Saved:      0,
		Skipped:    0,
		Duplicates: 0,
		Errors:     []string{},
	}

	toSave := s.filterDuplicates(ctx, data, opts, result)
	if len(toSave) == 0 {
		return result, nil
	}

	s.saveBatches(ctx, toSave, opts.BatchSize, result)
	result.Skipped = result.Total - result.Saved - result.Duplicates

	return result, nil
}

func (s *DataService) GetRecentData(ctx context.Context, configID string, limit int) ([]*models.FetchedData, error) {
	return s.fetchedDataRepo.GetByConfigID(ctx, configID, limit, 0)
}

func (s *DataService) GetDataSince(ctx context.Context, configID string, since time.Time, limit int) ([]*models.FetchedData, error) {
	return s.fetchedDataRepo.GetByConfigIDSince(ctx, configID, since, limit)
}

func (s *DataService) GetStats(ctx context.Context, configID string) (*DataStats, error) {
	count, err := s.fetchedDataRepo.Count(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to get count: %w", err)
	}

	recent, err := s.fetchedDataRepo.GetByConfigID(ctx, configID, 1, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent data: %w", err)
	}

	stats := &DataStats{
		TotalCount: count,
	}

	if len(recent) > 0 {
		stats.LastFetchedAt = &recent[0].FetchedAt
	}

	return stats, nil
}

func (s *DataService) filterDuplicates(ctx context.Context, data []*models.FetchedData, opts SaveOptions, result *SaveResult) []*models.FetchedData {
	var toSave []*models.FetchedData

	for _, item := range data {
		if item.ConfigID == "" && opts.ConfigID != "" {
			item.ConfigID = opts.ConfigID
		}

		if item.FetchedAt.IsZero() {
			item.FetchedAt = time.Now()
		}

		if s.isDuplicate(ctx, item, opts, result) {
			continue
		}

		toSave = append(toSave, item)
	}

	return toSave
}

func (s *DataService) isDuplicate(ctx context.Context, item *models.FetchedData, opts SaveOptions, result *SaveResult) bool {
	if opts.SkipDuplicatesByURL {
		exists, err := s.fetchedDataRepo.ExistsByURL(ctx, item.URL)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to check duplicate for URL %s: %v", item.URL, err))
			return true
		}
		if exists {
			result.Duplicates++
			return true
		}
	}

	if rawData, ok := item.RawData["id"]; ok {
		if redditID, ok := rawData.(string); ok {
			exists, err := s.fetchedDataRepo.ExistsByRedditID(ctx, redditID)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to check duplicate for Reddit ID %s: %v", redditID, err))
				return true
			}
			if exists {
				result.Duplicates++
				return true
			}
		}
	}

	return false
}

func (s *DataService) saveBatches(ctx context.Context, toSave []*models.FetchedData, batchSize int, result *SaveResult) {
	batches := s.splitIntoBatches(toSave, batchSize)
	for i, batch := range batches {
		if err := s.fetchedDataRepo.CreateBatch(ctx, batch); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to save batch %d: %v", i+1, err))
			continue
		}
		result.Saved += len(batch)
		log.Printf("Saved batch %d/%d (%d items)", i+1, len(batches), len(batch))
	}
}

func (s *DataService) splitIntoBatches(data []*models.FetchedData, batchSize int) [][]*models.FetchedData {
	var batches [][]*models.FetchedData
	for i := 0; i < len(data); i += batchSize {
		end := i + batchSize
		if end > len(data) {
			end = len(data)
		}
		batches = append(batches, data[i:end])
	}
	return batches
}

type SaveResult struct {
	Total      int      `json:"total"`
	Saved      int      `json:"saved"`
	Skipped    int      `json:"skipped"`
	Duplicates int      `json:"duplicates"`
	Errors     []string `json:"errors,omitempty"`
}

func (r *SaveResult) Success() bool {
	return len(r.Errors) == 0
}

func (r *SaveResult) Summary() string {
	if len(r.Errors) > 0 {
		return fmt.Sprintf("Total: %d, Saved: %d, Duplicates: %d, Skipped: %d, Errors: %d",
			r.Total, r.Saved, r.Duplicates, r.Skipped, len(r.Errors))
	}
	return fmt.Sprintf("Total: %d, Saved: %d, Duplicates: %d, Skipped: %d",
		r.Total, r.Saved, r.Duplicates, r.Skipped)
}

type DataStats struct {
	TotalCount     int64      `json:"total_count"`
	LastFetchedAt  *time.Time `json:"last_fetched_at,omitempty"`
}