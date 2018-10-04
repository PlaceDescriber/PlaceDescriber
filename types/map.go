package types

// map.go provides types for describing maps.

type MapType int

var MapTypesToStr = map[MapType]string{
	PLAN:       "plan",
	SATELLITE:  "satellite",
	HYBRID:     "hybrid",
	DESCRIPTOR: "descriptor",
}

const (
	PLAN MapType = iota
	SATELLITE
	HYBRID
	DESCRIPTOR
)
