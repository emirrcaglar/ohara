package db

import (
	"database/sql"
	"errors"
)

var ErrFolderCycle = errors.New("folder parent cannot be self or descendant")

type FolderRow struct {
	ID          int64
	ParentID    *int64
	Name        string
	ObjectCount int
}

func (db *DB) ListFolderChildren(parentID *int64) ([]FolderRow, error) {
	where := "parent_id IS NULL"
	args := []any{}
	if parentID != nil {
		where = "parent_id = ?"
		args = append(args, *parentID)
	}

	rows, err := db.Query(`
		SELECT c.id, c.parent_id, c.name,
		       (
		         SELECT COUNT(*) FROM folder child WHERE child.parent_id = c.id
		       ) + (
		         SELECT COUNT(*) FROM manga m WHERE m.folder_id = c.id
		       ) + (
		         SELECT COUNT(*) FROM audio a WHERE a.folder_id = c.id
		       ) + (
		         SELECT COUNT(*) FROM video v WHERE v.folder_id = c.id
		       ) AS object_count
		FROM folder c
		WHERE `+where+`
		ORDER BY c.name
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []FolderRow
	for rows.Next() {
		var row FolderRow
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

func (db *DB) ListFolderAll() ([]FolderRow, error) {
	rows, err := db.Query(`
		SELECT c.id, c.parent_id, c.name,
		       (
		         SELECT COUNT(*) FROM folder child WHERE child.parent_id = c.id
		       ) + (
		         SELECT COUNT(*) FROM manga m WHERE m.folder_id = c.id
		       ) + (
		         SELECT COUNT(*) FROM audio a WHERE a.folder_id = c.id
		       ) + (
		         SELECT COUNT(*) FROM video v WHERE v.folder_id = c.id
		       ) AS object_count
		FROM folder c
		ORDER BY c.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []FolderRow
	for rows.Next() {
		var row FolderRow
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

func (db *DB) GetFolderPath(id int64) ([]FolderRow, error) {
	rows, err := db.Query(`
		WITH RECURSIVE path(id, parent_id, name, depth) AS (
			SELECT id, parent_id, name, 0 FROM folder WHERE id = ?
			UNION ALL
			SELECT c.id, c.parent_id, c.name, path.depth + 1
			FROM folder c
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

	var path []FolderRow
	for rows.Next() {
		var row FolderRow
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

func (db *DB) InsertFolder(parentID *int64, name string) (*FolderRow, error) {
	var result sql.Result
	var err error
	if parentID == nil {
		result, err = db.Exec(`INSERT INTO folder (parent_id, name) VALUES (NULL, ?)`, name)
	} else {
		result, err = db.Exec(`INSERT INTO folder (parent_id, name) VALUES (?, ?)`, *parentID, name)
	}
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return db.GetFolderByID(id)
}

func (db *DB) GetFolderByID(id int64) (*FolderRow, error) {
	row := db.QueryRow(`
		SELECT c.id, c.parent_id, c.name,
		       (
		         SELECT COUNT(*) FROM folder child WHERE child.parent_id = c.id
		       ) + (
		         SELECT COUNT(*) FROM manga m WHERE m.folder_id = c.id
		       ) + (
		         SELECT COUNT(*) FROM audio a WHERE a.folder_id = c.id
		       ) + (
		         SELECT COUNT(*) FROM video v WHERE v.folder_id = c.id
		       ) AS object_count
		FROM folder c
		WHERE c.id = ?
	`, id)

	var folder FolderRow
	var parent sql.NullInt64
	if err := row.Scan(&folder.ID, &parent, &folder.Name, &folder.ObjectCount); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if parent.Valid {
		folder.ParentID = &parent.Int64
	}
	return &folder, nil
}

func (db *DB) UpdateFolder(id int64, parentID *int64, name string) (*FolderRow, error) {
	if parentID != nil {
		contains, err := db.folderContains(id, *parentID)
		if err != nil {
			return nil, err
		}
		if contains {
			return nil, ErrFolderCycle
		}
	}

	var result sql.Result
	var err error
	if parentID == nil {
		result, err = db.Exec(`UPDATE folder SET parent_id = NULL, name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, name, id)
	} else {
		result, err = db.Exec(`UPDATE folder SET parent_id = ?, name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, *parentID, name, id)
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

	return db.GetFolderByID(id)
}

func (db *DB) DeleteFolder(id int64) (bool, error) {
	tx, err := db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	// Remove contents before deleting folders: their foreign keys otherwise move
	// media and uploads to the library root via ON DELETE SET NULL.
	const descendants = `
		WITH RECURSIVE descendants(id) AS (
			SELECT id FROM folder WHERE id = ?
			UNION ALL
			SELECT f.id FROM folder f JOIN descendants d ON f.parent_id = d.id
		)
	`
	for _, query := range []string{
		`DELETE FROM manga_progress WHERE manga_id IN (SELECT id FROM manga WHERE folder_id IN (SELECT id FROM descendants))`,
		`DELETE FROM manga WHERE folder_id IN (SELECT id FROM descendants)`,
		`DELETE FROM audio WHERE folder_id IN (SELECT id FROM descendants)`,
		`DELETE FROM video_progress WHERE video_id IN (SELECT id FROM video WHERE folder_id IN (SELECT id FROM descendants))`,
		`DELETE FROM video WHERE folder_id IN (SELECT id FROM descendants)`,
		`DELETE FROM upload_sessions WHERE folder_id IN (SELECT id FROM descendants)`,
	} {
		if _, err := tx.Exec(descendants+query, id); err != nil {
			return false, err
		}
	}

	result, err := tx.Exec(`DELETE FROM folder WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

func (db *DB) folderContains(ancestorID, descendantID int64) (bool, error) {
	var exists int
	err := db.QueryRow(`
		WITH RECURSIVE descendants(id) AS (
			SELECT id FROM folder WHERE id = ?
			UNION ALL
			SELECT c.id FROM folder c
			JOIN descendants d ON c.parent_id = d.id
		)
		SELECT 1 FROM descendants WHERE id = ? LIMIT 1
	`, ancestorID, descendantID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}
