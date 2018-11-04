// tile.go: collection of algorithms and definitions for
// operating with coordinates and tiles.

package geography

import (
	"math"
	"time"

	"github.com/PlaceDescriber/PlaceDescriber/types"
)

const (
	D_R     = math.Pi / 180.0
	R_D     = 180.0 / math.Pi
	R_MAJOR = 6378137.0
	R_MINOR = 6356752.3142
	RATIO   = R_MINOR / R_MAJOR
)

var (
	ECCENT = math.Sqrt(1.0 - (RATIO * RATIO))
	COM    = 0.5 * ECCENT
)

type MapTile struct {
	Coordinates types.Point   `json:"coordinates"`
	Z           int           `json:"z"`
	Y           int           `json:"y"`
	X           int           `json:"x"`
	Time        time.Time     `json:"time"`
	Provider    string        `json:"provider"`
	Type        types.MapType `json:"type"`
	Language    string        `json:"language"`
	Content     []byte        `json:"content"`
}

type Conversion interface {
	DegToTileNum(coordinates types.Point, z int) (x, y int)
	TileNumToDeg(x, y, z int) types.Point
}

// Spherical Mercator.
// A popular spherical Mercator tiling format.
// This tile format is used by Google, OpenSteetMap and many others.
// It's based on a spherical Mercator projection ("Web Mercator", EPSG: 3857).
// http://wiki.openstreetmap.org/wiki/Slippy_map_tilenames
type SphericalConversion struct {
}

// DegToTileNum returns tile numbers by latitude and longitude in degrees.
func (s SphericalConversion) DegToTileNum(coordinates types.Point, z int) (x, y int) {
	x = int(math.Floor((coordinates.Longitude + 180.0) / 360.0 * (math.Exp2(float64(z)))))
	y = int(math.Floor((1.0 - math.Log(math.Tan(coordinates.Latitude*math.Pi/180.0)+1.0/math.Cos(coordinates.Latitude*math.Pi/180.0))/math.Pi) / 2.0 * (math.Exp2(float64(z)))))
	return
}

// TileNumToDeg returns latitude and longitude in degrees by tile numbers.
func (s SphericalConversion) TileNumToDeg(x, y, z int) types.Point {
	n := math.Pi - 2.0*math.Pi*float64(y)/math.Exp2(float64(z))
	lat := 180.0 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
	long := float64(x)/math.Exp2(float64(z))*360.0 - 180.0
	return types.Point{Latitude: lat, Longitude: long}
}

// Elliptical Mercator.
// This tile format is used at least by Yandex. Compared to the spherical
// format, the conversion between latitude/longitude coordinates and Mercator
// coordinates differs, but the tiling logic is otherwise the same.
// http://wiki.openstreetmap.org/wiki/Mercator#Elliptical_Mercator
type EllipticalConversion struct {
}

// DegToTileNum returns tile numbers by latitude and longitude in degrees.
func (s EllipticalConversion) DegToTileNum(coordinates types.Point, z int) (x, y int) {
	xmerc := D_R * coordinates.Longitude
	lat := math.Min(89.5, math.Max(coordinates.Latitude, -89.5))
	phi := D_R * lat
	sinphi := math.Sin(phi)
	con := ECCENT * sinphi
	con = math.Pow((1.0-con)/(1.0+con), COM)
	ts := math.Tan(0.5*(math.Pi*0.5-phi)) / con
	ymerc := -math.Log(ts)
	x = int((1 + xmerc/math.Pi) / 2 * math.Exp2(float64(z)))
	y = int((1 - ymerc/math.Pi) / 2 * math.Exp2(float64(z)))
	return
}

// TileNumToDeg returns latitude and longitude in degrees by tile numbers.
func (s EllipticalConversion) TileNumToDeg(x, y, z int) types.Point {
	xmerc := (float64(x)/math.Exp2(float64(z))*2 - 1) * math.Pi
	ymerc := (1 - float64(y)/math.Exp2(float64(z))*2) * math.Pi
	long := R_D * xmerc
	ts := math.Exp(-ymerc)
	phi := math.Pi/2 - 2*math.Atan(ts)
	dphi := 1.0
	for i := 0; math.Abs(dphi) > 0.000000001 && i < 15; i++ {
		con := ECCENT * math.Sin(phi)
		dphi = math.Pi/2 - 2*math.Atan(ts*math.Pow((1.0-con)/(1.0+con), COM)) - phi
		phi += dphi
	}
	lat := R_D * phi
	return types.Point{Latitude: lat, Longitude: long}
}
