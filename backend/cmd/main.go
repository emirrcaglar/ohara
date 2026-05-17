package main

import (
	"flag"
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
		log.Error("[main] failed to init database err=%v", err)
		return
	}
	log.Info("[main] database initialized data_dir=%s", *dataDir)
	defer database.Close()

	// Bootstrap admin user
	if *domain == "" {
		if err := database.EnsureAdmin("admin", "admin"); err != nil {
			log.Error("[main] failed to bootstrap local admin user err=%v", err)
		} else {
			log.Info("[main] local admin bootstrap complete username=admin")
		}
	} else {
		adminUser := os.Getenv("OHARA_ADMIN_USER")
		adminPass := os.Getenv("OHARA_ADMIN_PASS")
		if adminUser != "" && adminPass != "" {
			if err := database.EnsureAdmin(adminUser, adminPass); err != nil {
				log.Error("[main] failed to bootstrap admin user err=%v", err)
			} else {
				log.Info("[main] admin bootstrap complete username=%s", adminUser)
			}
		}
	}

	if *scan != "" {
		if flag.NArg() == 0 {
			log.Warn("[scan] missing directory for type=%s", *scan)
			return
		}
		dir := flag.Arg(0)
		log.Info("[scan] requested type=%s dir=%s", *scan, dir)
		s := scanner.NewScanner(database, cbz.NewCBZService(database), log, scanner.WithScanDir(dir), scanner.WithScanType(scanner.ScanType(*scan)))
		scannedCount, err := s.Run()
		if err != nil {
			log.Error("[scan] failed type=%s dir=%s err=%v", *scan, dir, err)
			return
		}
		log.Info("[scan] indexed count=%d type=%s dir=%s", scannedCount, *scan, dir)
		return
	}

	r := router.SetupRoutes(database, *dataDir, log)

	log.Info("[main] Ohara listening on port %s", *port)

	if err := server.Start(server.Config{Domain: *domain, Port: *port, DataDir: *dataDir}, r, log); err != nil {
		log.Error("Server crashed: %v", err)
	}
}
