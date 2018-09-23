package types

// map.go provides types for describing maps.

type MapType int

const (
	PLAN MapType = iota
	SATELLITE
	HYBRID
	DESCRIPTOR
)
