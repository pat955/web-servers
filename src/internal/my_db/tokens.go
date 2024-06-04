package my_db

import "time"

type TokenInfo struct {
	UserID     int       `json:"user_id"`
	ExpiresUTC time.Time `json:"expires_utc"`
}

func (db *DB) GetRefreshToken(token string) (TokenInfo, bool) {
	data := db.loadDB()
	f, ok := data.RefreshTokens[token]
	return f, ok
}

func (db *DB) Revoke(token string) error {
	data := db.loadDB()
	delete(data.RefreshTokens, token)
	db.writeDB(data)
	return nil
}
