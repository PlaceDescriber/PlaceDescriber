package mapget

import (
	"bytes"
	"context"
	"errors"
	"image/png"
	"io"
	"io/ioutil"
	"sync"
	"testing"

	"github.com/PlaceDescriber/PlaceDescriber/geography"
	"github.com/PlaceDescriber/PlaceDescriber/types"
)

const (
	// Downloading parameters.
	GOROUTINES_NUMBER = 10
	TRY_TIMES         = 5
	// RGB bounds for content test.
	RED_MIN   = 42000
	RED_MAX   = 53000
	GREEN_MIN = 49000
	GREEN_MAX = 59000
	BLUE_MIN  = 53000
	BLUE_MAX  = 63000
)

type TestLoader struct {
	UrlsToFail  int
	FailsPerUrl int
	currFails   map[string]int
	mtx         sync.Mutex
}

func (s *TestLoader) Do(ctx context.Context, url string) (io.ReadCloser, error) {
	s.mtx.Lock()
	if len(s.currFails) < s.UrlsToFail {
		if _, ok := s.currFails[url]; !ok {
			s.currFails[url] = 0
		}
	}
	if fails, ok := s.currFails[url]; ok {
		if fails < s.FailsPerUrl {
			s.currFails[url] = fails + 1
			s.mtx.Unlock()
			return nil, errors.New("Error.")
		}
	}
	s.mtx.Unlock()
	return ioutil.NopCloser(bytes.NewBufferString(url)), nil
}

func newTestLoader(urlsToFail int, failsPerUrl int) *TestLoader {
	return &TestLoader{
		UrlsToFail:  urlsToFail,
		FailsPerUrl: failsPerUrl,
		currFails:   make(map[string]int),
	}
}

func initMapDescription() MapDescription {
	// For now the values are ContentTest-oriented, since
	// they don't matter for the rest of the tests.
	area := types.Polygon{
		Vertices: []types.Point{
			// The Atlantic Ocean.
			types.Point{40.710476, -22.466572},
			types.Point{40.716877, -22.453697},
			types.Point{40.715669, -22.451766},
			types.Point{40.721644, -22.432003},
			types.Point{40.719889, -22.428441},
			types.Point{40.716869, -22.437904},
		},
	}
	return MapDescription{
		MapArea:  area,
		Provider: "yandex",
		Type:     types.PLAN,
		Language: "ru_RU",
		MinZoom:  14,
		MaxZoom:  19,
		Scale:    1,
	}
}

func TestIncorrectInput(t *testing.T) {
	var wg sync.WaitGroup
	var err error
	out := make(chan *geography.MapTile)
	loader := newTestLoader(1, 1)
	params := DownloadParams{
		GoroutinesNum: 10001,
		TryTimes:      0,
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = DownloadMap(
			context.Background(),
			params,
			initMapDescription(),
			out,
			loader,
		)
	}()
	for _ = range out {
	}
	wg.Wait()
	if err == nil {
		t.Errorf("DownloadMap didn't fail on incorrect input.")
	}
}

func TestSingleFailure(t *testing.T) {
	var wg sync.WaitGroup
	var err error
	out := make(chan *geography.MapTile)
	loader := newTestLoader(1, 1)
	params := DownloadParams{
		GoroutinesNum: GOROUTINES_NUMBER,
		TryTimes:      TRY_TIMES,
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = DownloadMap(
			context.Background(),
			params,
			initMapDescription(),
			out,
			loader,
		)
	}()
	for _ = range out {
	}
	wg.Wait()
	if err != nil {
		t.Errorf("DownloadMap failed on single loading error %v.", err)
	}
}

func TestMultipleFailures(t *testing.T) {
	var wg sync.WaitGroup
	var err error
	out := make(chan *geography.MapTile)
	loader := newTestLoader(1, TRY_TIMES)
	params := DownloadParams{
		GoroutinesNum: GOROUTINES_NUMBER,
		TryTimes:      TRY_TIMES,
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = DownloadMap(
			context.Background(),
			params,
			initMapDescription(),
			out,
			loader,
		)
	}()
	for _ = range out {
	}
	wg.Wait()
	if err == nil {
		t.Errorf("DownloadMap did not fail on multiple loading errors.")
	}
}

func TestCancel(t *testing.T) {
	var wg sync.WaitGroup
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	out := make(chan *geography.MapTile)
	loader := newTestLoader(0, 0)
	params := DownloadParams{
		GoroutinesNum: GOROUTINES_NUMBER,
		TryTimes:      TRY_TIMES,
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = DownloadMap(
			ctx,
			params,
			initMapDescription(),
			out,
			loader,
		)
	}()
	cancel()
	for _ = range out {
	}
	wg.Wait()
	if err == nil {
		t.Errorf("DownloadMap did not fail on context cancellation.")
	}
}

func TestSuccess(t *testing.T) {
	var wg sync.WaitGroup
	var err error
	out := make(chan *geography.MapTile)
	loader := newTestLoader(0, 0)
	params := DownloadParams{
		GoroutinesNum: GOROUTINES_NUMBER,
		TryTimes:      TRY_TIMES,
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = DownloadMap(
			context.Background(),
			params,
			initMapDescription(),
			out,
			loader,
		)
	}()
	for _ = range out {
	}
	wg.Wait()
	if err != nil {
		t.Errorf("DownloadMap failed without loading problems: %v.", err)
	}
}

func TestContent(t *testing.T) {
	var wg sync.WaitGroup
	var err error
	out := make(chan *geography.MapTile)
	loader := DefaultLoader{}
	params := DownloadParams{
		GoroutinesNum: GOROUTINES_NUMBER,
		TryTimes:      TRY_TIMES,
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = DownloadMap(
			context.Background(),
			params,
			initMapDescription(),
			out,
			loader,
		)
	}()
	for tile := range out {
		img, err1 := png.Decode(bytes.NewReader(tile.Content))
		if err1 != nil {
			t.Fatalf("ContentTest: failed to decode image: %v", err1)
		}
		x := img.Bounds().Max.X / 2
		y := img.Bounds().Max.Y / 2
		r, g, b, _ := img.At(x, y).RGBA()
		// Ocean must be blue.
		if r < RED_MIN || r > RED_MAX {
			t.Errorf("ContentTest: bad RGB for ocean tiles: red %d", r)
		}
		if g < GREEN_MIN || g > GREEN_MAX {
			t.Errorf("ContentTest: bad RGB for ocean tiles: green %d.", g)
		}
		if b < BLUE_MIN || b > BLUE_MAX {
			t.Errorf("ContentTest: bad RGB for ocean tiles: blue %d.", b)
		}
	}
	wg.Wait()
	if err != nil {
		t.Errorf("ContentTest: DownloadMap failed: %v.", err)
	}
}
