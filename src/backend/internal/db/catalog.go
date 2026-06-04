package db

import (
	"database/sql"
	"errors"
)

var ErrCatalogCycle = errors.New("catalog parent cannot be self or descendant")

type CatalogRow struct {
	ID          int64
	ParentID    *int64
	Name        string
	ObjectCount int
}

func (db *DB) ListCatalogChildren(parentID *int64) ([]CatalogRow, error) {
	where := "parent_id IS NULL"
	args := []any{}
	if parentID != nil {
		where = "parent_id = ?"
		args = append(args, *parentID)
	}

	rows, err := db.Query(`
		SELECT c.id, c.parent_id, c.name,
		       (
		         SELECT COUNT(*) FROM catalog child WHERE child.parent_id = c.id
		       ) + (
		         SELECT COUNT(*) FROM manga m WHERE m.catalog_id = c.id
		       ) + (
		         SELECT COUNT(*) FROM audio a WHERE a.catalog_id = c.id
		       ) + (
		         SELECT COUNT(*) FROM video v WHERE v.catalog_id = c.id
		       ) AS object_count
		FROM catalog c
		WHERE `+where+`
		ORDER BY c.name
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []CatalogRow
	for rows.Next() {
		var row CatalogRow
		var parent sql.NullInt64
		if err := rows.Scan(&row.ID, &parent, &row.Name, &row.ObjectCount); err != nil {
			return nil, err
		}
		if parent.Valid {
			row.ParentID = &parent.Int64
		}
		list = append(list, row)
	}
	return list, rows.Err()
}

func (db *DB) ListCatalogAll() ([]CatalogRow, error) {
	rows, err := db.Query(`
		SELECT c.id, c.parent_id, c.name,
		       (
		         SELECT COUNT(*) FROM catalog child WHERE child.parent_id = c.id
		       ) + (
		         SELECT COUNT(*) FROM manga m WHERE m.catalog_id = c.id
		       ) + (
		         SELECT COUNT(*) FROM audio a WHERE a.catalog_id = c.id
		       ) + (
		         SELECT COUNT(*) FROM video v WHERE v.catalog_id = c.id
		       ) AS object_count
		FROM catalog c
		ORDER BY c.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []CatalogRow
	for rows.Next() {
		var row CatalogRow
		var parent sql.NullInt64
		if err := rows.Scan(&row.ID, &parent, &row.Name, &row.ObjectCount); err != nil {
			return nil, err
		}
		if parent.Valid {
			row.ParentID = &parent.Int64
		}
		list = append(list, row)
	}
	return list, rows.Err()
}

func (db *DB) GetCatalogPath(id int64) ([]CatalogRow, error) {
	rows, err := db.Query(`
		WITH RECURSIVE path(id, parent_id, name, depth) AS (
			SELECT id, parent_id, name, 0 FROM catalog WHERE id = ?
			UNION ALL
			SELECT c.id, c.parent_id, c.name, path.depth + 1
			FROM catalog c
			JOIN path ON path.parent_id = c.id
		)
		SELECT id, parent_id, name, 0 AS object_count
		FROM path
		ORDER BY depth DESC
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var path []CatalogRow
	for rows.Next() {
		var row CatalogRow
		var parent sql.NullInt64
		if err := rows.Scan(&row.ID, &parent, &row.Name, &row.ObjectCount); err != nil {
			return nil, err
		}
		if parent.Valid {
			row.ParentID = &parent.Int64
		}
		path = append(path, row)
	}
	return path, rows.Err()
}

func (db *DB) InsertCatalog(parentID *int64, name string) (*CatalogRow, error) {
	var result sql.Result
	var err error
	if parentID == nil {
		result, err = db.Exec(`INSERT INTO catalog (parent_id, name) VALUES (NULL, ?)`, name)
	} else {
		result, err = db.Exec(`INSERT INTO catalog (parent_id, name) VALUES (?, ?)`, *parentID, name)
	}
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return db.GetCatalogByID(id)
}

func (db *DB) GetCatalogByID(id int64) (*CatalogRow, error) {
	row := db.QueryRow(`
		SELECT c.id, c.parent_id, c.name,
		       (
		         SELECT COUNT(*) FROM catalog child WHERE child.parent_id = c.id
		       ) + (
		         SELECT COUNT(*) FROM manga m WHERE m.catalog_id = c.id
		       ) + (
		         SELECT COUNT(*) FROM audio a WHERE a.catalog_id = c.id
		       ) + (
		         SELECT COUNT(*) FROM video v WHERE v.catalog_id = c.id
		       ) AS object_count
		FROM catalog c
		WHERE c.id = ?
	`, id)

	var catalog CatalogRow
	var parent sql.NullInt64
	if err := row.Scan(&catalog.ID, &parent, &catalog.Name, &catalog.ObjectCount); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if parent.Valid {
		catalog.ParentID = &parent.Int64
	}
	return &catalog, nil
}

func (db *DB) UpdateCatalog(id int64, parentID *int64, name string) (*CatalogRow, error) {
	if parentID != nil {
		contains, err := db.catalogContains(id, *parentID)
		if err != nil {
			return nil, err
		}
		if contains {
			return nil, ErrCatalogCycle
		}
	}

	var result sql.Result
	var err error
	if parentID == nil {
		result, err = db.Exec(`UPDATE catalog SET parent_id = NULL, name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, name, id)
	} else {
		result, err = db.Exec(`UPDATE catalog SET parent_id = ?, name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, *parentID, name, id)
	}
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, nil
	}

	return db.GetCatalogByID(id)
}

func (db *DB) DeleteCatalog(id int64) (bool, error) {
	result, err := db.Exec(`DELETE FROM catalog WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

func (db *DB) catalogContains(ancestorID, descendantID int64) (bool, error) {
	var exists int
	err := db.QueryRow(`
		WITH RECURSIVE descendants(id) AS (
			SELECT id FROM catalog WHERE id = ?
			UNION ALL
			SELECT c.id FROM catalog c
			JOIN descendants d ON c.parent_id = d.id
		)
		SELECT 1 FROM descendants WHERE id = ? LIMIT 1
	`, ancestorID, descendantID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}
