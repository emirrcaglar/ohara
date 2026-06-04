package db

import (
	"database/sql"
	"time"
)

type Deployment struct {
	ID         int       `json:"id"`
	DeployedAt time.Time `json:"deployedAt"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (db *DB) RecordDeployment(deployedAt time.Time) error {
	_, err := db.Exec(
		`INSERT OR IGNORE INTO deployments (deployed_at) VALUES (?)`,
		deployedAt.UTC().Format(time.RFC3339),
	)
	return err
}

func (db *DB) GetLatestDeployment() (*Deployment, error) {
	row := db.QueryRow(`
		SELECT id, deployed_at, created_at
		FROM deployments
		ORDER BY deployed_at DESC
		LIMIT 1
	`)

	var deployment Deployment
	var deployedAt string
	var createdAt string
	if err := row.Scan(&deployment.ID, &deployedAt, &createdAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	parsedDeployedAt, err := time.Parse(time.RFC3339, deployedAt)
	if err != nil {
		return nil, err
	}
	deployment.DeployedAt = parsedDeployedAt

	parsedCreatedAt, err := time.Parse("2006-01-02 15:04:05", createdAt)
	if err != nil {
		deployment.CreatedAt = deployment.DeployedAt
	} else {
		deployment.CreatedAt = parsedCreatedAt.UTC()
	}

	return &deployment, nil
}
