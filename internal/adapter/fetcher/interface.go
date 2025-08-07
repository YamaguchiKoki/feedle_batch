package fetcher

import (
	"context"

	"github.com/YamaguchiKoki/feedle_batch/internal/domain/model"
)

type Fetcher[T model.FetchConfigDetail] interface {
	Name() string
	Fetch(ctx context.Context, config T) ([]*model.FetchedData, error)
}
