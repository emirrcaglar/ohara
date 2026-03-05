package main

import (
	"flag"
	"fmt"
	"log"

	"ohara/src/internal/db"
	"ohara/src/internal/router"
	"ohara/src/internal/server"
)

func main() {
	domain := flag.String("domain", "", "Domain for auto-HTTPS (e.g., stream.example.com)")
	port := flag.String("port", "8080", "Local dev port")
	dataDir := flag.String("data", "./app-data", "Path to store certs and media")
	mangaDir := flag.String("manga", ".", "Path to manga directory")
	flag.Parse()

	cfg := server.Config{
		Domain:  *domain,
		Port:    *port,
		DataDir: *dataDir,
	}

	database, err := db.Init(*dataDir)
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	defer database.Close()

	r := router.SetupRoutes(*mangaDir, database)

	fmt.Printf("Ohara port: %s\n", *port)
	fmt.Printf("Manga base dir: %s\n", *mangaDir)

	if err := server.Start(cfg, r); err != nil {
		log.Fatalf("Server crashed: %v", err)
	}
}
