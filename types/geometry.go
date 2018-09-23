package types

// geometry.go defines some basic geometric types
// in the context of geography and maps.

import (
	"errors"
	"sort"
)

type (
	Point struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	// Polygon is an outline based on the vertex list.
	Polygon struct {
		Vertices []Point `json:"vertices"`
	}

	// Custom types for sorting.
	ByLatitude  []Point
	ByLongitude []Point
)

func (s ByLatitude) Len() int {
	return len(s)
}

func (s ByLatitude) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByLatitude) Less(i, j int) bool {
	return s[i].Latitude < s[j].Latitude
}

func (s ByLongitude) Len() int {
	return len(s)
}

func (s ByLongitude) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByLongitude) Less(i, j int) bool {
	return s[i].Longitude < s[j].Longitude
}

// ExtremeCoordinates returns minimum and maximum values for
// latitude and longitude.
func (p Polygon) ExtremeCoordinates() (minLat, minLong, maxLat, maxLong float64, err error) {
	size := len(p.Vertices)
	if size == 0 {
		err = errors.New("Trying to apply ExtremeCoordinates to Polygon of 0 points.")
	}
	sort.Sort(ByLatitude(p.Vertices))
	minLat, maxLat = p.Vertices[0].Latitude, p.Vertices[size-1].Latitude
	sort.Sort(ByLongitude(p.Vertices))
	minLong, maxLong = p.Vertices[0].Longitude, p.Vertices[size-1].Longitude
	return
}
