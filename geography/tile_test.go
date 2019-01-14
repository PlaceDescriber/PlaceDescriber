package geography

import (
	"errors"
	"math"
	"testing"

	"github.com/PlaceDescriber/PlaceDescriber/types"
)

const (
	ZOOM                      = 14
	LATITUDE                  = 55.269203
	LONGITUDE                 = 36.841963
	SPHERICAL_TILE_LATITUDE   = 55.279115
	SPHERICAL_TILE_LONGITUDE  = 36.826172
	ELLIPTICAL_TILE_LATITUDE  = 55.271208
	ELLIPTICAL_TILE_LONGITUDE = 36.826172
	SPHERICAL_X               = 9868
	SPHERICAL_Y               = 5160
	ELLITPTICAL_X             = 9868
	ELLIPTICAL_Y              = 5175
)

func testConversion(converter Conversion, x, y int, lat, long float64) error {
	res_x, res_y := converter.DegToTileNum(
		types.Point{Latitude: LATITUDE, Longitude: LONGITUDE},
		ZOOM,
	)
	if res_x != x {
		return errors.New("DegToTileNum returned invalid x")
	}
	if res_y != y {
		return errors.New("DegToTileNum returned invalid y")
	}
	coordinates := converter.TileNumToDeg(x, y, ZOOM)
	if math.Abs(coordinates.Latitude-lat) > 1e-6 {
		return errors.New("TileNumToDeg returned invalid latitude")
	}
	if math.Abs(coordinates.Longitude-long) > 1e-6 {
		return errors.New("TileNumToDeg returned invalid longitude")
	}
	return nil
}

func TestSphericalConversion(t *testing.T) {
	converter := SphericalConversion{}
	err := testConversion(
		converter,
		SPHERICAL_X,
		SPHERICAL_Y,
		SPHERICAL_TILE_LATITUDE,
		SPHERICAL_TILE_LONGITUDE,
	)
	if err != nil {
		t.Errorf("SphericalConversion: %v.", err)
	}
}

func TestEllipticalConversion(t *testing.T) {
	converter := EllipticalConversion{}
	err := testConversion(
		converter,
		ELLITPTICAL_X,
		ELLIPTICAL_Y,
		ELLIPTICAL_TILE_LATITUDE,
		ELLIPTICAL_TILE_LONGITUDE,
	)
	if err != nil {
		t.Errorf("EllipticalConversion: %v.", err)
	}
}
