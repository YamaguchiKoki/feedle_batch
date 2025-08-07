package output

import (
	"context"

	"github.com/YamaguchiKoki/feedle_batch/internal/domain/model"
)

// UserRepository はUserのリポジトリインターフェース
type UserRepository interface {
	GetActiveUserIDs(ctx context.Context) ([]model.UserID, error)
}
