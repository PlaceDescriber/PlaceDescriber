package mapget

// mapget.go provides map downloader implementation.

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/PlaceDescriber/PlaceDescriber/geography"
	"github.com/PlaceDescriber/PlaceDescriber/types"
	"golang.org/x/sync/semaphore"
)

const (
	GOROUTINES_MAX = 50
)

// TODO: think if we need JSON config file and flag for that.
type TypeToUrl map[types.MapType]string

// TODO: support more map providers.
var URLS = map[string]TypeToUrl{
	"yandex": TypeToUrl{
		types.PLAN:      "https://vec01.maps.yandex.net/tiles?l=map&x=%s&y=%s&z=%s&scale=%s&lang=%s",
		types.SATELLITE: "https://sat01.maps.yandex.net/tiles?l=sat&x=%s&y=%s&z=%s&scale=%s&lang=%s",
	},
}

type MapDescription struct {
	MapArea  types.Polygon `json:"map_area":`
	Provider string        `json:"provider"`
	Type     types.MapType `json:"type"`
	Language string        `json:"language"`
	MinZoom  int           `json:"min_zoom"`
	MaxZoom  int           `json:"max_zoom"`
	Scale    int           `json:"scale"`
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func extremeTileNumbers(z int, mapArea types.Polygon) (int, int, int, int, error) {
	minLat, minLong, maxLat, maxLong, err := mapArea.ExtremeCoordinates()
	x0, y0 := geography.DegToTileNum(types.Point{minLat, minLong}, z)
	x1, y1 := geography.DegToTileNum(types.Point{maxLat, maxLong}, z)
	return min(x0, x1), max(x0, x1), min(y0, y1), max(y0, y1), err
}

func prepareHeader(header *http.Header) {
	// TODO: check if these headers are relevant, check the order.
	header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; rv:52.0) Gecko/20100101 Firefox/52.0")
	header.Set("Accept", "*/*")
	header.Set("Accept-Language", "en-US,en;q=0.5")
	header.Set("Accept-Encoding", "gzip, deflate, br")
	header.Set("Referer", "blablabla")
	header.Set("Connection", "keep-alive")
}

func downloadTile(
	tile *geography.MapTile,
	scale int,
	out chan *geography.MapTile,
	err chan error,
) {
	client := &http.Client{}
	_, ok := URLS[tile.Provider]
	if !ok {
		err <- errors.New(fmt.Sprintf("downloadTile: bad map provider %s", tile.Provider))
		return
	}
	url, ok := URLS[tile.Provider][tile.Type]
	if !ok {
		err <- errors.New(fmt.Sprintf("downloadTile: bad map type %s", tile.Type))
		return
	}
	fmt.Sprintf(url, tile.X, tile.Y, tile.Z, scale, tile.Language)
	req, err0 := http.NewRequest("GET", url, nil)
	if err0 != nil {
		err <- err0
		return
	}
	prepareHeader(&req.Header)
	res, err0 := client.Do(req)
	if err0 != nil {
		err <- err0
		return
	}
	defer res.Body.Close()
	tile.Content, err0 = ioutil.ReadAll(res.Body)
	if err0 != nil {
		err <- err0
		return
	}
	out <- tile
}

func DownloadMap(mapDesc MapDescription, out chan *geography.MapTile, err chan error) {
	var wg sync.WaitGroup
	ctx := context.TODO()
	sem := semaphore.NewWeighted(int64(GOROUTINES_MAX))
	for z := mapDesc.MinZoom; z <= mapDesc.MaxZoom; z++ {
		minX, maxX, minY, maxY, err0 := extremeTileNumbers(z, mapDesc.MapArea)
		if err0 != nil {
			err <- err0
			return
		}
		for x := minX; x <= maxX; x++ {
			for y := minY; y <= maxY; y++ {
				if err1 := sem.Acquire(ctx, 1); err1 != nil {
					err <- errors.New(fmt.Sprintf("Failed to acquire semaphore: %v", err1))
					return
				}
				tile := &geography.MapTile{
					Z:        z,
					Y:        y,
					X:        x,
					Time:     time.Now(),
					Provider: mapDesc.Provider,
					Type:     mapDesc.Type,
					Language: mapDesc.Language,
				}
				wg.Add(1)
				go func() {
					defer sem.Release(1)
					defer wg.Done()
					downloadTile(tile, mapDesc.Scale, out, err)
				}()
			}
		}
	}
	wg.Wait()
	err <- nil
	close(out)
	close(err)
}
