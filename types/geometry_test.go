package types

import (
	"testing"
)

func TestExtremeCoordinates(t *testing.T) {
	p := Polygon{
		Vertices: []Point{
			Point{30.12300, 67.32320},
			Point{40.21410, 20.32434},
			Point{35.43423, 88.3242340},
		},
	}
	minLat, minLong, maxLat, maxLong, err := p.ExtremeCoordinates()
	if err != nil {
		t.Fatalf("ExtremeCoordinates: %v.", err)
	}
	if minLat != 30.12300 {
		t.Fatalf("ExtremeCoordinates: invalid minimum latitude.")
	}
	if minLong != 20.32434 {
		t.Fatalf("ExtremeCoordinates: invalid minimum longitude.")
	}
	if maxLat != 40.21410 {
		t.Fatalf("ExtremeCoordinates: invalid maximum latitude.")
	}
	if maxLong != 88.3242340 {
		t.Fatalf("ExtremeCoordinates: invalid maximum longitude.")
	}
}
