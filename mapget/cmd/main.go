package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/PlaceDescriber/PlaceDescriber/geography"
	"github.com/PlaceDescriber/PlaceDescriber/mapget"
)

var (
	mapConf       = flag.String("map-conf", "", "Path to JSON file describing what to download.")
	downloadDir   = flag.String("download-dir", "~/maps", "Directory for tiles.")
	provider      = flag.String("provider", "yandex", "One of possible map providers.")
	retryTimes    = flag.Int("retry-times", 5, "Number of tries to download each of the tiles.")
	goroutinesNum = flag.Int("goroutines-num", 10, "Number of goroutines to use while loading.")
)

type MapConfig struct {
	MapName string
	MapDesc mapget.MapDescription
}

func main() {
	flag.Parse()
	file, err := os.Open(*mapConf)
	defer file.Close()
	if err != nil {
		log.Fatalf("Can't open map config file: %v", err)
	}
	var mapConf MapConfig
	data, _ := ioutil.ReadAll(file)
	json.Unmarshal(data, &mapConf)
	out := make(chan *geography.MapTile)
	params := mapget.DownloadParams{
		GoroutinesNum: *goroutinesNum,
		RetryTimes:    *retryTimes,
	}
	go func() {
		err = mapget.DownloadMap(
			context.Background(),
			params,
			mapConf.MapDesc,
			out,
			mapget.DefaultLoader{},
		)
	}()
	for _ = range out {

	}
	if err != nil {
		log.Fatalf("DownloadMap: %v", err)
	}
}
