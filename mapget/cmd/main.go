package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"sync"
	"time"

	"github.com/PlaceDescriber/PlaceDescriber/geography"
	"github.com/PlaceDescriber/PlaceDescriber/mapget"
	"github.com/PlaceDescriber/PlaceDescriber/types"
)

var (
	mapName       = flag.String("map-name", "", "Map name, will be used as sub-dir for tiles.")
	coordinates   = flag.String("coordinates", "", "Path to JSON file describing what to download.")
	provider      = flag.String("provider", "yandex", "One of possible map providers.")
	mapType       = flag.Int("map-type", int(types.SATELLITE), "Map type.")
	language      = flag.String("language", "en_EN", "Map language.")
	minZoom       = flag.Int("min-zoom", 14, "Zoom level to start with.")
	maxZoom       = flag.Int("max-zoom", 19, "Zoom level to finish with.")
	scale         = flag.Int("scale", 1, "Tiles scale.")
	downloadDir   = flag.String("download-dir", "~/maps", "Directory for tiles.")
	goroutinesNum = flag.Int("goroutines-num", 10, "Number of goroutines to use while loading.")
	retryTimes    = flag.Int("retry-times", 5, "Number of tries to download each of the tiles.")
)

const PATH_TEMPLATE = "%s/%s/%s/%s/%s/%s"

func getCurTime() string {
	now := time.Now()
	return fmt.Sprintf("%d-%d-%d", now.Year(), now.Month(), now.Day())
}

func expandTilde(path string) (string, error) {
	if path[0] != '~' {
		return path, nil
	}
	if path[1] != '/' {
		return "", fmt.Errorf("path %q starts with '~' but not with '~/'", path)
	}
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("user.Current: %s", err)
	}
	return filepath.Join(usr.HomeDir, path[2:]), nil
}

func makeDir(path string) error {
	wantMode := os.FileMode(0700 | os.ModeDir)
	_ = os.MkdirAll(path, wantMode)
	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("os.Stat(%s): %s", path, err)
	}
	if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}
	if stat.Mode() != wantMode {
		return fmt.Errorf("%s has mode %s, want %s", path, stat.Mode(), wantMode)
	}
	return nil
}
func main() {
	flag.Parse()
	if len(*mapName) == 0 {
		log.Fatalf("You must specify map name with map-name option.")
	}
	if len(*coordinates) == 0 {
		log.Fatalf("You must specify path to JSON file with coordinates using coordinates option.")
	}
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	mapDesc := mapget.MapDescription{
		Provider: *provider,
		Type:     types.MapType(*mapType),
		Language: *language,
		MinZoom:  *minZoom,
		MaxZoom:  *maxZoom,
		Scale:    *scale,
	}
	coordinatesFile, err := os.Open(*coordinates)
	defer coordinatesFile.Close()
	if err != nil {
		log.Fatalf("Can't open map coordinates file: %v.", err)
	}
	data, err := ioutil.ReadAll(coordinatesFile)
	if err != nil {
		log.Fatalf("Can't read map coordinates file: %v.", err)
	}
	json.Unmarshal(data, &mapDesc.MapArea)
	out := make(chan *geography.MapTile)
	params := mapget.DownloadParams{
		GoroutinesNum: *goroutinesNum,
		RetryTimes:    *retryTimes,
	}
	path := fmt.Sprintf(
		PATH_TEMPLATE,
		*downloadDir,
		*mapName,
		*provider,
		types.MapTypesToStr[types.MapType(*mapType)],
		*language,
		getCurTime(),
	)
	path, err = expandTilde(path)
	if err != nil {
		log.Fatalf("Failed to expand ~ to home dir in path: %v.", err)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = mapget.DownloadMap(
			ctx,
			params,
			mapDesc,
			out,
			mapget.DefaultLoader{},
		)
	}()
	for tile := range out {
		tileDirPath := fmt.Sprintf("%s/%d/%d/", path, tile.Z, tile.X)
		if err0 := makeDir(tileDirPath); err0 != nil {
			log.Fatalf("Failed to make/check tile dir: %v.", err)
		}
		tileFileName := fmt.Sprintf("%d", tile.Y)
		tileFile, err0 := os.Create(filepath.Join(tileDirPath, tileFileName))
		if err0 != nil {
			cancel()
			wg.Wait()
			log.Fatalf("Failed to creare tile file: %v", err0)
		}
		_, err0 = tileFile.Write(tile.Content)
		if err0 != nil {
			cancel()
			wg.Wait()
			log.Fatalf("Failed to write to tile file: %v", err0)
		}
		tileFile.Close()
	}
	wg.Wait()
	if err != nil {
		log.Fatalf("DownloadMap: %v.", err)
	}
}
