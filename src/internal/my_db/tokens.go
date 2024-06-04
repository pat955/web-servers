package my_db

import "fmt"

func (db *DB) GetToken(token string) RefreshTokenInfo {
	data, err := db.loadDB()
	if err != nil {
		panic(err)
	}
	return data.RefreshToken[token]
}

func (db *DB) Revoke(token string) error {
	data, err := db.loadDB()
	if err != nil {
		return err
	}
	fmt.Println(len(data.RefreshToken))

	delete(data.RefreshToken, token)
	fmt.Println(len(data.RefreshToken))
	return nil
}
