package main

import (
	"flag"
	"os"
	"time"

	"ohara/src/internal/db"
	"ohara/src/internal/logger"
	"ohara/src/internal/router"
	"ohara/src/internal/server"
)

func main() {
	domain := flag.String("domain", "", "Domain for auto-HTTPS (e.g., stream.example.com)")
	port := flag.String("port", "3000", "Local dev port")
	dataDir := flag.String("data", "./app-data", "Path to store certs and media")
	flag.Parse()

	log := logger.New(500)

	database, err := db.Init(*dataDir)
	if err != nil {
		log.Error("[main] failed to init database err=%v", err)
		return
	}
	log.Info("[main] database initialized data_dir=%s", *dataDir)
	defer database.Close()

	if deployedAt := os.Getenv("OHARA_DEPLOYED_AT"); deployedAt != "" {
		parsedDeployedAt, err := time.Parse(time.RFC3339, deployedAt)
		if err != nil {
			log.Error("[main] invalid OHARA_DEPLOYED_AT value=%s err=%v", deployedAt, err)
		} else if err := database.RecordDeployment(parsedDeployedAt); err != nil {
			log.Error("[main] failed to record deployment deployed_at=%s err=%v", deployedAt, err)
		} else {
			log.Info("[main] deployment recorded deployed_at=%s", deployedAt)
		}
	}

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

	r := router.SetupRoutes(database, *dataDir, log)

	log.Info("[main] Ohara listening on port %s", *port)

	if err := server.Start(server.Config{Domain: *domain, Port: *port, DataDir: *dataDir}, r, log); err != nil {
		log.Error("Server crashed: %v", err)
	}
}
