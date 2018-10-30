package types

// map.go provides types for describing maps.

type MapType int

var StrToMapType = map[string]MapType{
	"plan":       PLAN,
	"satellite":  SATELLITE,
	"hybrid":     HYBRID,
	"descriptor": DESCRIPTOR,
}

var MapTypeToStr map[MapType]string

const (
	PLAN MapType = iota
	SATELLITE
	HYBRID
	DESCRIPTOR
)

func init() {
	MapTypeToStr = make(map[MapType]string)
	for key, val := range StrToMapType {
		MapTypeToStr[val] = key
	}
}
