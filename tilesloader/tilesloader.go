// tilesloader calls mapget to get tiles, saves them
// to the tile database if needed and publishes them to
// pubsub.

package tilesloader

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/NYTimes/gizmo/pubsub"
	"github.com/PlaceDescriber/PlaceDescriber/databases"
	"github.com/PlaceDescriber/PlaceDescriber/geography"
	"github.com/PlaceDescriber/PlaceDescriber/mapget"
)

type TilesHandler struct {
	publisher      pubsub.Publisher
	client         databases.TilesDatabase
	downloadParams mapget.DownloadParams
}

func (s *TilesHandler) ServeWithSave(w http.ResponseWriter, r *http.Request) {
	s.serveTiles(w, r, true)
}

func (s *TilesHandler) ServeOnlyHandling(w http.ResponseWriter, r *http.Request) {
	s.serveTiles(w, r, false)
}

func (s *TilesHandler) serveTiles(w http.ResponseWriter, r *http.Request, save bool) {
	var mapDesc mapget.MapDescription
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&mapDesc)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	tiles := make(chan *geography.MapTile)
	go mapget.DownloadMap(
		ctx,
		s.downloadParams,
		mapDesc,
		tiles,
		mapget.DefaultLoader{},
	)
	for tile := range tiles {
		if save {
			_, err := s.client.SaveTile(ctx, tile)
			if err != nil {
				http.Error(w, err.Error(), 500)
				cancel()
				return
			}
		}
		tileBytes, err := json.Marshal(tile)
		if err != nil {
			http.Error(w, err.Error(), 500)
			cancel()
			return
		}
		err = s.publisher.PublishRaw(nil, "", tileBytes)
		if err != nil {
			http.Error(w, err.Error(), 500)
			cancel()
			return
		}
	}
}

func NewHandler(
	downloadParams mapget.DownloadParams,
	publisher pubsub.Publisher,
	client databases.TilesDatabase,
) http.Handler {
	tilesHandler := TilesHandler{
		downloadParams: downloadParams,
		publisher:      publisher,
		client:         client,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/load-tiles/save", tilesHandler.ServeWithSave)
	mux.HandleFunc("/load-tiles/handle", tilesHandler.ServeOnlyHandling)
	return mux
}
