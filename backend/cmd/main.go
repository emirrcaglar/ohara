package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"ohara/src/internal/db"
	"ohara/src/internal/logger"
	"ohara/src/internal/media/cbz"
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

	log := logger.New(500)

	cacheDir := filepath.Join(*dataDir, "cache")
	go worker.StartCacheCleaner(cacheDir, 1000, 15*time.Minute, log) // 1 GB

	database, err := db.Init(*dataDir)
	if err != nil {
		log.Error("Failed to init database: %v", err)
		return
	}
	defer database.Close()

	// Bootstrap admin user
	if *domain == "" {
		if err := database.EnsureAdmin("admin", "admin"); err != nil {
			log.Error("Failed to bootstrap local admin user: %v", err)
		}
	} else {
		adminUser := os.Getenv("OHARA_ADMIN_USER")
		adminPass := os.Getenv("OHARA_ADMIN_PASS")
		if adminUser != "" && adminPass != "" {
			if err := database.EnsureAdmin(adminUser, adminPass); err != nil {
				log.Error("Failed to bootstrap admin user: %v", err)
			}
		}
	}

	if *scan != "" {
		if flag.NArg() == 0 {
			fmt.Println("Usage: --scan <type> <directory>")
			return
		}
		dir := flag.Arg(0)
		s := scanner.NewScanner(database, cbz.NewCBZService(database), scanner.WithScanDir(dir), scanner.WithScanType(scanner.ScanType(*scan)))
		scannedCount, err := s.Run()
		if err != nil {
			fmt.Printf("Scan failed: %v", err)
		}
		fmt.Printf("Indexed %d new %s from %s\n", scannedCount, *scan, dir)
		return
	}

	r := router.SetupRoutes(database, *dataDir, log)

	log.Info("Ohara listening on port %s", *port)

	if err := server.Start(server.Config{Domain: *domain, Port: *port, DataDir: *dataDir}, r); err != nil {
		log.Error("Server crashed: %v", err)
	}
}
