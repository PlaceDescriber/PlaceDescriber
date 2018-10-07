package mapget

import (
	"fmt"

	"github.com/PlaceDescriber/PlaceDescriber/geography"
	"github.com/PlaceDescriber/PlaceDescriber/types"
)

type TypeToUrl map[types.MapType]string

type MapProject interface {
	Converter() geography.Conversion
	GetURL(x, y, z, scale int, language string, mapType types.MapType) (string, error)
}

// TODO: support more map providers.
var MapProjects = map[string]MapProject{
	"yandex": YandexMaps{},
}

type YandexMaps struct {
}

func (s YandexMaps) Converter() geography.Conversion {
	return geography.EllipticalConversion{}
}

func (s YandexMaps) GetURL(
	x, y, z, scale int,
	language string,
	mapType types.MapType,
) (string, error) {
	urls := TypeToUrl{
		types.PLAN:      "https://vec01.maps.yandex.net/tiles?l=map&x=%d&y=%d&z=%d&scale=%d&lang=%s",
		types.SATELLITE: "https://sat01.maps.yandex.net/tiles?l=sat&x=%d&y=%d&z=%d&scale=%d&lang=%s",
	}
	url, ok := urls[mapType]
	if !ok {
		return "", fmt.Errorf("YandexMaps doesn't support map type %d", mapType)
	}
	return fmt.Sprintf(url, x, y, z, scale, language), nil
}
