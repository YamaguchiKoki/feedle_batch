package output

import (
	"context"

	"github.com/YamaguchiKoki/feedle_batch/internal/domain/model"
)

type DataSourceRepository interface {
	GetByID(ctx context.Context, id string) (*model.DataSource, error)
	GetAll(ctx context.Context) ([]*model.DataSource, error)
	GetActive(ctx context.Context) ([]*model.DataSource, error)
}
