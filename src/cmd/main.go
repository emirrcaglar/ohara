package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"ohara/src/internal/db"
	"ohara/src/internal/indexer"
	"ohara/src/internal/router"
	"ohara/src/internal/server"
	"ohara/src/internal/worker"
)

func main() {
	domain := flag.String("domain", "", "Domain for auto-HTTPS (e.g., stream.example.com)")
	port := flag.String("port", "8080", "Local dev port")
	dataDir := flag.String("data", "./app-data", "Path to store certs and media")
	scanDir := flag.String("scan-manga", "", "Scan a directory for manga and exit")
	flag.Parse()

	cacheDir := filepath.Join(*dataDir, "cache")
	fmt.Printf("[cache] Cache Directory: %s", cacheDir)
	// 1 GB
	go worker.StartCacheCleaner(cacheDir, 1000, 15*time.Minute)

	database, err := db.Init(*dataDir)
	if err != nil {
		fmt.Printf("Failed to init database: %v", err)
	}
	defer database.Close()

	if *scanDir != "" {
		added, err := indexer.Run(database, *scanDir)
		if err != nil {
			fmt.Printf("Scan failed: %v", err)
		}
		fmt.Printf("Indexed %d new manga from %s\n", added, *scanDir)
		return
	}

	r := router.SetupRoutes(database, *dataDir)

	fmt.Printf("Ohara port: %s\n", *port)

	if err := server.Start(server.Config{Domain: *domain, Port: *port, DataDir: *dataDir}, r); err != nil {
		fmt.Printf("Server crashed: %v", err)
	}
}
