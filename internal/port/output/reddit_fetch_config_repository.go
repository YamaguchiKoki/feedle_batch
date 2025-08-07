package output

import (
	"context"

	"github.com/YamaguchiKoki/feedle_batch/internal/domain/model"
	"github.com/google/uuid"
)

type RedditFetchConfigRepository interface {
	GetByUserFetchConfigID(ctx context.Context, userFetchConfigID uuid.UUID) (*model.RedditFetchConfig, error)
}
