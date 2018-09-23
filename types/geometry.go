package types

// geometry.go defines some basic geometric types
// in the context of geography and maps.

type (
	Point struct {
		Latitude float64  `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	// Polygon is an outline based on the vertex list.
	Polygon struct {
		Vertices []Point `json:"vertices"`
	}

)
