package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v4"
	seeker "github.com/malcolmmadsheep/handshakes-seeker/cmd/seeker/app"
	"github.com/malcolmmadsheep/handshakes-seeker/internal/dbhandlers"
	"github.com/malcolmmadsheep/handshakes-seeker/internal/dbservices"
	"github.com/malcolmmadsheep/handshakes-seeker/pkg/plugin"
	"github.com/malcolmmadsheep/handshakes-seeker/plugins"
)

func main() {
	cfg := seeker.Config{}

	log.Println("Connecting to database...")
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Couldn't set up connection with database %s", err)
	}
	log.Println("Successfully connected to database...")

	err = runDBMigration(conn)
	if err != nil {
		log.Fatalf("DB migration failed %s", err)
	}

	taskService := dbservices.NewTaskService(conn)
	handlers := dbhandlers.New(conn, taskService)
	wikipediaPlugin := &plugins.WikipediaPlugin{}

	plugins := []plugin.Plugin{wikipediaPlugin}

	skr, err := seeker.New(context.Background(), cfg, handlers, taskService, plugins)
	if err != nil {
		log.Fatalf("Failed to run seeker: %s", err)
	}

	if err := skr.Run(); err != nil {
		log.Fatalf("Seeker is shutdown")
	}
}
