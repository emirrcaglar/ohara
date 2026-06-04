package main

import (
	"os"
	"time"

	"ohara/src/internal/db"
	"ohara/src/internal/logger"
	"ohara/src/internal/router"
	"ohara/src/internal/server"
)

const dataDir = "app-data"

func main() {
	log := logger.New(500)

	database, err := db.Init(dataDir)
	if err != nil {
		log.Error("[main] failed to init database err=%v", err)
		return
	}
	log.Info("[main] database initialized data_dir=%s", dataDir)
	if database.DefaultAdminCreated() {
		log.Info("[main] default admin account created username=admin password=admin; change this password immediately")
	}
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

	r := router.SetupRoutes(database, dataDir, log)

	if err := server.Start(r, log); err != nil {
		log.Error("Server crashed: %v", err)
	}
}
