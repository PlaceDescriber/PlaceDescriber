// databases package provides abstraction from certain
// databases used in the project.

package databases

import (
	"github.com/PlaceDescriber/PlaceDescriber/geography"
)

type TilesDatabase interface {
	SaveTile(*geography.MapTile) (string, error)
	GetTile(key string) (*geography.MapTile, error)
}

type KeyValueStorage interface {
	Set(bucket, key string, value []byte) error
	Get(bucket, key string) ([]byte, error)
}
