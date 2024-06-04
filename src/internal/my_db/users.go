package my_db

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type PublicUser struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

type UserTokenResponse struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func (u *User) UserLoginResponse() UserTokenResponse {
	return UserTokenResponse{ID: u.ID, Email: u.Email, Token: u.AccessToken, RefreshToken: u.RefreshToken}
}

func (u *User) UserToPublic() PublicUser {
	return PublicUser{ID: u.ID, Email: u.Email}
}

func (u *User) GenerateClaims() *jwt.RegisteredClaims {
	// 1h
	expires := time.Now().UTC().Add(time.Second * time.Duration(3600))
	claims := &jwt.RegisteredClaims{
		Issuer:    "Chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(expires),
		Subject:   fmt.Sprint(u.ID),
	}
	// if u.ExpiresInSeconds > 0 && u.ExpiresInSeconds < 86400 {
	// 	claims.ExpiresAt = jwt.NewNumericDate(time.Now().UTC().Add(time.Second * time.Duration(u.ExpiresInSeconds)))
	// }
	return claims
}

func (u *User) GenerateToken() string {
	jwtSecret := os.Getenv("JWT_SECRET")
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, u.GenerateClaims())
	token, err := t.SignedString([]byte(jwtSecret))
	if err != nil {
		panic(err)
	}
	return token
}

func (u *User) GenerateRefreshToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("error:", err)
		return ""
	}
	return hex.EncodeToString(b)
}

func GeneratePassword(pass string) []byte {
	passByte, err := bcrypt.GenerateFromPassword([]byte(pass), 10)
	if err != nil {
		panic(err)
	}
	return passByte
}

func (db *DB) AddUser(user User) {
	data, err := db.loadDB()
	if err != nil {
		panic(err)
	}
	user.Password = string(GeneratePassword(user.Password))
	user.RefreshToken = user.GenerateRefreshToken()
	user.AccessToken = user.GenerateToken()

	data.RefreshTokens[user.RefreshToken] = TokenInfo{UserID: user.ID, ExpiresUTC: time.Now().UTC().Add(time.Hour * time.Duration(1440))}
	data.Users[user.ID] = user
	db.writeDB(data)
}

func (db *DB) UpdateUser(user User) {
	data, err := db.loadDB()
	if err != nil {
		panic(err)
	}
	foundUser := data.Users[user.ID]

	if user.Password == foundUser.Password {
	} else if err := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(user.Password)); err == nil {
		user.Password = foundUser.Password
	} else {
		user.Password = string(GeneratePassword(user.Password))
	}
	data.Users[user.ID] = user
	db.writeDB(data)
}

func (db *DB) GetUsers() []User {
	db.mux.RLock()
	f, err := os.ReadFile(db.Path)
	if err != nil {
		panic(err)
	}
	db.mux.RUnlock()
	var data DBStructure
	json.Unmarshal(f, &data)
	var allUsers []User
	for _, user := range data.Users {
		allUsers = append(allUsers, user)
	}
	return allUsers
}

func (db *DB) GetUser(id int) (User, bool) {
	data, err := db.loadDB()
	if err != nil {
		panic(err)
	}
	user, found := data.Users[id]
	if !found {
		return User{}, false
	}
	return user, true
}
