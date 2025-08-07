package output

import (
	"context"

	"github.com/YamaguchiKoki/feedle_batch/internal/domain/model"
)

type FetchedDataRepository interface {
	Create(ctx context.Context, data *model.FetchedData) error
}
