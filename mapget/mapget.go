// Package mapget provides map downloader implementation.
package mapget

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/PlaceDescriber/PlaceDescriber/geography"
	"github.com/PlaceDescriber/PlaceDescriber/types"
	"golang.org/x/net/context/ctxhttp"
)

const (
	TILE_SIZE      = 256
	MIN_ZOOM       = 0
	MAX_ZOOM       = 25
	MIN_SCALE      = 0
	MAX_SCALE      = 5
	MIN_TRY_TIMES  = 1
	MAX_TRY_TIMES  = 20
	MIN_GOROUTINES = 1
	MAX_GOROUTINES = 1000
)

type MapDescription struct {
	MapArea  types.Polygon `json:"map_area"`
	Provider string        `json:"provider"`
	Type     types.MapType `json:"type"`
	Language string        `json:"language"`
	MinZoom  int           `json:"min_zoom"`
	MaxZoom  int           `json:"max_zoom"`
	Scale    int           `json:"scale"`
}

type DownloadParams struct {
	GoroutinesNum int `json:"goroutines_num"`
	TryTimes      int `json:"try_times"`
}

type DownloadTask struct {
	Tile  *geography.MapTile `json:"tile"`
	Scale int                `json:"scale"`
}

type Loader interface {
	Do(ctx context.Context, url string) (io.ReadCloser, error)
}

type DefaultLoader struct {
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

func (s DefaultLoader) Do(ctx context.Context, url string) (io.ReadCloser, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	prepareHeader(&req.Header)
	res, err := ctxhttp.Do(ctx, client, req)
	if err != nil {
		return nil, err
	}
	return res.Body, nil
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

func checkInput(mapDesc MapDescription, params DownloadParams) error {
	if params.TryTimes < MIN_TRY_TIMES || params.TryTimes > MAX_TRY_TIMES {
		return fmt.Errorf("TryTimes is out of range: %d", params.TryTimes)
	}
	if params.GoroutinesNum < MIN_GOROUTINES || params.GoroutinesNum > MAX_GOROUTINES {
		return fmt.Errorf("GoroutinesNum is out of range: %d", params.GoroutinesNum)
	}
	_, ok := MapProjects[mapDesc.Provider]
	if !ok {
		return fmt.Errorf("Bad map provider %s", mapDesc.Provider)
	}
	typeStr, ok := types.MapTypeToStr[mapDesc.Type]
	if !ok {
		return fmt.Errorf("Bad map type %s", typeStr)
	}
	// TODO: check language here.
	if mapDesc.MinZoom < MIN_ZOOM || mapDesc.MinZoom > MAX_ZOOM {
		return fmt.Errorf("MinZoom is out of range: %d", mapDesc.MinZoom)
	}
	if mapDesc.MaxZoom < MIN_ZOOM || mapDesc.MaxZoom > MAX_ZOOM {
		return fmt.Errorf("MaxZoom is out of range: %d", mapDesc.MaxZoom)
	}
	if mapDesc.Scale < MIN_SCALE || mapDesc.Scale > MAX_SCALE {
		return fmt.Errorf("Scale is out of range: %d", mapDesc.Scale)
	}
	if mapDesc.MinZoom >= mapDesc.MaxZoom {
		return fmt.Errorf("MinZoom is greater or equal to MaxZoom")
	}
	return nil
}

func extremeTileNumbers(
	z int,
	mapArea types.Polygon,
	converter geography.Conversion,
) (int, int, int, int, error) {
	minLat, minLong, maxLat, maxLong, err := mapArea.ExtremeCoordinates()
	x0, y0 := converter.DegToTileNum(
		types.Point{Latitude: minLat, Longitude: minLong},
		z,
	)
	x1, y1 := converter.DegToTileNum(
		types.Point{Latitude: maxLat, Longitude: maxLong},
		z,
	)
	return min(x0, x1), max(x0, x1), min(y0, y1), max(y0, y1), err
}

func downloadTile(
	ctx context.Context,
	task *DownloadTask,
	client Loader,
) (*geography.MapTile, error) {
	tile := task.Tile
	mapProj, ok := MapProjects[tile.Provider]
	if !ok {
		return nil, fmt.Errorf("downloadTile: bad map provider %s", tile.Provider)
	}
	url, err := mapProj.GetURL(tile.X, tile.Y, tile.Z, task.Scale, tile.Language, tile.Type)
	if err != nil {
		return nil, err
	}
	body, err := client.Do(ctx, url)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	tile.Content, err = ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	return tile, nil
}

func downloadTileWrapper(
	ctx context.Context,
	tryTimes int,
	task *DownloadTask,
	client Loader,
) (*geography.MapTile, error) {
	for i := 0; i < tryTimes; i++ {
		tile, err := downloadTile(ctx, task, client)
		if err == nil {
			return tile, nil
		}
		log.Printf("downloadTile failed with %v", err)
	}
	err := fmt.Errorf("downloadTileWrapper: tried %d times and failed", tryTimes)
	log.Printf("%v", err)
	return nil, err
}

func solveTasks(
	ctx context.Context,
	tryTimes int,
	tasks <-chan *DownloadTask,
	out chan<- *geography.MapTile,
	client Loader,
) error {
	for task := range tasks {
		tile, err := downloadTileWrapper(ctx, tryTimes, task, client)
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- tile:
		}
	}
	return nil
}

func createTasks(
	ctx context.Context,
	mapDesc MapDescription,
	tasks chan<- *DownloadTask,
) error {
	for z := mapDesc.MinZoom; z <= mapDesc.MaxZoom; z++ {
		mapProj, ok := MapProjects[mapDesc.Provider]
		if !ok {
			close(tasks)
			return fmt.Errorf("createTasks: bad map provider %s", mapDesc.Provider)
		}
		minX, maxX, minY, maxY, err := extremeTileNumbers(z, mapDesc.MapArea, mapProj.Converter())
		if err != nil {
			close(tasks)
			return err
		}
		for x := minX; x <= maxX; x++ {
			for y := minY; y <= maxY; y++ {
				tile := &geography.MapTile{
					Z:        z,
					Y:        y,
					X:        x,
					Time:     time.Now(),
					Provider: mapDesc.Provider,
					Type:     mapDesc.Type,
					Language: mapDesc.Language,
				}
				task := &DownloadTask{
					Tile:  tile,
					Scale: mapDesc.Scale,
				}
				select {
				case <-ctx.Done():
					close(tasks)
					return ctx.Err()
				case tasks <- task:
				}
			}
		}
	}
	close(tasks)
	return nil
}

func DownloadMap(
	ctx context.Context,
	params DownloadParams,
	mapDesc MapDescription,
	out chan<- *geography.MapTile,
	client Loader,
) error {
	err := checkInput(mapDesc, params)
	if err != nil {
		close(out)
		log.Printf("Incorrect input: %v.\n", err)
		return err
	}
	ctx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	var mtx sync.Mutex
	tasks := make(chan *DownloadTask)
	wg.Add(1)
	go func() {
		defer wg.Done()
		err1 := createTasks(ctx, mapDesc, tasks)
		if err1 != nil {
			log.Printf("Task creation failed with %v.\n", err1)
			cancel()
			mtx.Lock()
			err = err1
			mtx.Unlock()
		}
	}()
	for i := 0; i < params.GoroutinesNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err1 := solveTasks(ctx, params.TryTimes, tasks, out, client)
			if err1 != nil {
				log.Printf("Task failed with %v.\n", err1)
				mtx.Lock()
				err = err1
				mtx.Unlock()
				cancel()
			}
		}()
	}
	wg.Wait()
	close(out)
	return err
}
