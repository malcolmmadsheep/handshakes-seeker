package main

import (
	"log"

	seeker "github.com/malcolmmadsheep/handshakes-seeker/cmd/seeker/app"
	"github.com/malcolmmadsheep/handshakes-seeker/pkg/plugin"
	"github.com/malcolmmadsheep/handshakes-seeker/plugins"
)

func main() {
	cfg := seeker.Config{}

	wikipediaPlugin := &plugins.WikipediaPlugin{}

	plugins := []plugin.Plugin{wikipediaPlugin}

	skr, err := seeker.New(cfg, plugins)
	if err != nil {
		log.Fatalf("Failed to run seeker: %s", err)
	}

	if err := skr.Run(); err != nil {
		log.Fatalf("Seeker is shutdown")
	}
}
