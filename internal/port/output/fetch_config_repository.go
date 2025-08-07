package output

import (
	"context"

	"github.com/YamaguchiKoki/feedle_batch/internal/domain/model"
)

type FetchConfigRepository interface {
	GetByUserID(ctx context.Context, userID model.UserID) ([]*model.UserFetchConfig, error)
}
