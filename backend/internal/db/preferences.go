package db

import "database/sql"

type PreferenceRow struct {
	Key   string
	Value string
}

func (db *DB) ListPreferences(userID int64) (map[string]string, error) {
	rows, err := db.Query(`SELECT key, value FROM preferences WHERE user_id = ? ORDER BY key`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	preferences := make(map[string]string)
	for rows.Next() {
		var p PreferenceRow
		if err := rows.Scan(&p.Key, &p.Value); err != nil {
			return nil, err
		}
		preferences[p.Key] = p.Value
	}
	return preferences, rows.Err()
}

func (db *DB) GetPreference(userID int64, key string) (string, bool, error) {
	var value string
	err := db.QueryRow(`SELECT value FROM preferences WHERE user_id = ? AND key = ?`, userID, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return value, true, nil
}

func (db *DB) UpsertPreference(userID int64, key, value string) error {
	_, err := db.Exec(`
		INSERT INTO preferences (user_id, key, value, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
	`, userID, key, value)
	return err
}

func (db *DB) DeletePreference(userID int64, key string) error {
	_, err := db.Exec(`DELETE FROM preferences WHERE user_id = ? AND key = ?`, userID, key)
	return err
}
