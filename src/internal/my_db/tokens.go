package my_db

func (db *DB) GetRefreshToken(token string) (TokenInfo, bool) {
	data, err := db.loadDB()
	if err != nil {
		panic(err)
	}
	f, ok := data.RefreshTokens[token]
	return f, ok
}

func (db *DB) Revoke(token string) error {
	data, err := db.loadDB()
	if err != nil {
		return err
	}
	delete(data.RefreshTokens, token)
	db.writeDB(data)
	return nil
}
