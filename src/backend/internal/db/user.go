package db

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Role         string
	IsApproved   bool
	PFP          int
}

func (db *DB) GetUserByUsername(username string) (*User, error) {
	row := db.QueryRow(`
		SELECT id, username, password_hash, role, is_approved, pfp
		FROM user WHERE username = ?
	`, username)

	var u User
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.IsApproved, &u.PFP)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (db *DB) UpdateUserPassword(userID int64, passwordHash string) error {
	_, err := db.Exec(`
		UPDATE user SET password_hash = ?
		WHERE id = ?
	`, passwordHash, userID)
	return err
}

func (db *DB) UpdateUserPFP(userID int64, pfp int) error {
	_, err := db.Exec(`
		UPDATE user SET pfp = ?
		WHERE id = ?
	`, pfp, userID)
	return err
}
