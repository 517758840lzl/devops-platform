package main

import (
	"log"

	"devops-platform/internal/config"
	"devops-platform/internal/database"
	"devops-platform/internal/router"
)

func main() {
	cfg := config.Load()
	db, err := database.Connect(cfg.DBPath)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if err := database.Seed(db); err != nil {
		log.Fatalf("seed: %v", err)
	}

	r := router.New(db, cfg)
	log.Printf("server listening on http://localhost:%s data=%s", cfg.Port, cfg.DataDir)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
