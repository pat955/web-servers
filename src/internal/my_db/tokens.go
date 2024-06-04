package my_db

import "fmt"

func (db *DB) GetRefreshToken(token string) TokenInfo {
	data, err := db.loadDB()
	if err != nil {
		panic(err)
	}
	return data.RefreshTokens[token]
}

func (db *DB) Revoke(token string) error {
	data, err := db.loadDB()
	if err != nil {
		return err
	}
	fmt.Println(len(data.RefreshTokens))

	delete(data.RefreshTokens, token)
	fmt.Println(len(data.RefreshTokens))
	return nil
}
