//go:build wails

package main

import (
	"context"
	"embed"
	"log"

	"bot_nardy/internal/desktop"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	api, err := desktop.NewAPI("moves.jsonl")
	if err != nil {
		log.Fatalf("desktop init failed: %v", err)
	}

	err = wails.Run(&options.App{
		Title:  "Desktop Nardy Engine",
		Width:  1320,
		Height: 900,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 248, G: 244, B: 233, A: 1},
		OnShutdown: func(ctx context.Context) {
			_ = api.Close()
		},
		Bind: []interface{}{api},
	})
	if err != nil {
		log.Fatal(err)
	}
}
