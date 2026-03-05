package main

import (
	"flag"
	"fmt"
	"log"

	"ohara/src/internal/db"
	"ohara/src/internal/indexer"
	"ohara/src/internal/router"
	"ohara/src/internal/server"
)

func main() {
	domain := flag.String("domain", "", "Domain for auto-HTTPS (e.g., stream.example.com)")
	port := flag.String("port", "8080", "Local dev port")
	dataDir := flag.String("data", "./app-data", "Path to store certs and media")
	scanDir := flag.String("scan-manga", "", "Scan a directory for manga and exit")
	flag.Parse()

	database, err := db.Init(*dataDir)
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	defer database.Close()

	if *scanDir != "" {
		added, err := indexer.Run(database, *scanDir)
		if err != nil {
			log.Fatalf("Scan failed: %v", err)
		}
		fmt.Printf("Indexed %d new manga from %s\n", added, *scanDir)
		return
	}

	r := router.SetupRoutes(database)

	fmt.Printf("Ohara port: %s\n", *port)

	if err := server.Start(server.Config{Domain: *domain, Port: *port, DataDir: *dataDir}, r); err != nil {
		log.Fatalf("Server crashed: %v", err)
	}
}
