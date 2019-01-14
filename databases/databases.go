// databases package provides abstraction from certain
// databases used in the project.

package databases

import (
	"context"

	"github.com/PlaceDescriber/PlaceDescriber/geography"
)

type TilesDatabase interface {
	SaveTile(ctx context.Context, tile *geography.MapTile) (string, error)
	GetTile(ctx context.Context, key string) (*geography.MapTile, error)
}

type KeyValueStorage interface {
	Set(ctx context.Context, bucket, key string, value []byte) error
	Get(ctx context.Context, bucket, key string) ([]byte, error)
}
