package geography

// tile.go: collection of algorithms and definitions for
// operating with coordinates and tiles.

import (
	"math"
	"time"

	"github.com/PlaceDescriber/PlaceDescriber/types"
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

func DegToTileNum(coordinates types.Point, z int) (x, y int) {
	x = int(math.Floor((coordinates.Longitude + 180.0) / 360.0 * (math.Exp2(float64(z)))))
	y = int(math.Floor((1.0 - math.Log(math.Tan(coordinates.Latitude*math.Pi/180.0)+1.0/math.Cos(coordinates.Latitude*math.Pi/180.0))/math.Pi) / 2.0 * (math.Exp2(float64(z)))))
	return
}

func TileNumToDeg(x, y, z int) types.Point {
	n := math.Pi - 2.0*math.Pi*float64(y)/math.Exp2(float64(z))
	lat := 180.0 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
	long := float64(x)/math.Exp2(float64(z))*360.0 - 180.0
	return types.Point{lat, long}
}
