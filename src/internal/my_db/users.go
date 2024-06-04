package my_db

import (
	"crypto/rand"
	"encoding/hex"
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

func (u *User) UserToPublic() PublicUser {
	return PublicUser{ID: u.ID, Email: u.Email}
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

// Generates jwt claims with 1 hour expiration. Subject is the users id.
func (u *User) GenerateClaims() *jwt.RegisteredClaims {
	return &jwt.RegisteredClaims{
		Issuer:    "Chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Second * time.Duration(3600))),
		Subject:   fmt.Sprint(u.ID),
	}
}

// Adds new user and refresh token to db
func (db *DB) AddUser(user User) {
	data := db.loadDB()
	user.Password = string(GeneratePassword(user.Password))
	user.RefreshToken = user.GenerateRefreshToken()
	user.AccessToken = user.GenerateToken()

	data.RefreshTokens[user.RefreshToken] = TokenInfo{
		UserID:     user.ID,
		ExpiresUTC: time.Now().UTC().Add(time.Hour * time.Duration(1440)),
	}
	data.Users[user.ID] = user
	db.writeDB(data)
}

// checks if password is already encrypted
func (db *DB) UpdateUser(user User) {
	data := db.loadDB()
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

// use GetUser when you can for preformance
func (db *DB) GetUsers() []User {
	data := db.loadDB()
	var allUsers []User
	for _, user := range data.Users {
		allUsers = append(allUsers, user)
	}
	return allUsers
}

func (db *DB) GetUser(id int) (User, bool) {
	data := db.loadDB()

	user, found := data.Users[id]
	if !found {
		return User{}, false
	}
	return user, true
}

// -------------------------------------------------

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
