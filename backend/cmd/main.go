package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"ohara/src/internal/db"
	"ohara/src/internal/router"
	"ohara/src/internal/scanner"
	"ohara/src/internal/server"
	"ohara/src/internal/worker"
)

func main() {
	domain := flag.String("domain", "", "Domain for auto-HTTPS (e.g., stream.example.com)")
	port := flag.String("port", "3000", "Local dev port")
	dataDir := flag.String("data", "./app-data", "Path to store certs and media")
	scan := flag.String("scan", "", "Scan for media: all, manga, or audio")
	flag.Parse()

	cacheDir := filepath.Join(*dataDir, "cache")
	go worker.StartCacheCleaner(cacheDir, 1000, 15*time.Minute) // 1 GB

	database, err := db.Init(*dataDir)
	if err != nil {
		fmt.Printf("Failed to init database: %v\n", err)
		return
	}
	defer database.Close()

	if *scan != "" {
		if flag.NArg() == 0 {
			fmt.Println("Usage: --scan <type> <directory>")
			return
		}
		dir := flag.Arg(0)
		s := scanner.NewScanner(database, dir, scanner.ScanType(*scan))
		scannedCount, err := s.Run()
		if err != nil {
			fmt.Printf("Scan failed: %v", err)
		}
		fmt.Printf("Indexed %d new %s from %s\n", scannedCount, *scan, dir)
		return
	}

	r := router.SetupRoutes(database, *dataDir)

	fmt.Printf("Ohara port: %s\n", *port)

	if err := server.Start(server.Config{Domain: *domain, Port: *port, DataDir: *dataDir}, r); err != nil {
		fmt.Printf("Server crashed: %v", err)
	}
}
