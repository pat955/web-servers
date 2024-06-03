package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"net/http"
	"regexp"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

const DBPATH string = "./database.json"

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("err loading: %v", err)
	}
	jwtSecret := os.Getenv("JWT_SECRET")

	// debug flag, deletes the db if $ ./out --debug
	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	if *dbg {
		deleteDB(DBPATH)
	}

	const root = "../public"
	const port = "8080"
	apiCfg := apiConfig{
		fileserverHits: 0,
		JWTSecret:      jwtSecret,
	}

	router := mux.NewRouter()
	defaultHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(root))))
	router.Handle("/app/*", middlewareLog(defaultHandler))
	router.HandleFunc("/admin/metrics", apiCfg.handlerCount).Methods("GET")
	router.HandleFunc("/api/healthz", handlerStatus).Methods("GET")
	router.HandleFunc("/api/chirps", handlerAddChirp).Methods("POST")
	router.HandleFunc("/api/chirps", handlerGetChirps).Methods("GET")
	router.HandleFunc("/api/chirps/{chirpID}", handlerAddChirpId).Methods("GET")
	router.HandleFunc("/api/users", handlerAddUser).Methods("POST")
	router.HandleFunc("/api/users", handlerAuth).Methods("PUT")
	router.HandleFunc("/api/login", handlerLogin).Methods("POST")
	router.HandleFunc("/api/reset", apiCfg.handlerResetCount)
	corsMux := middlewareLog(middlewareCors(router))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}
	srv.ListenAndServe()
}

func handlerAuth(w http.ResponseWriter, req *http.Request) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		respondWithError(w, 500, "JWT secret not set")
		return
	}

	db, err := createDB(DBPATH)
	if err != nil {
		panic(err)
	}
	auth := req.Header.Get("Authorization")
	if auth == "Bearer: " {
		respondWithError(w, 401, "Authorization header missing")
		return
	}

	tokenString := strings.Split(auth, "Bearer ")
	if len(tokenString) != 2 {
		respondWithError(w, 401, "Invalid Authorization header format")
		return
	}

	token, err := jwt.ParseWithClaims(tokenString[1], &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		respondWithError(w, 401, "Unauthorized response, "+err.Error())
		return
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		respondWithError(w, 401, "Invalid token")
		return
	}

	email := claims.Subject
	if email == "" {
		respondWithError(w, 401, "Token subject is missing")
		return
	}
	users := db.getUsersMap()
	foundUser, found := users[email]
	if !found {
		respondWithError(w, 404, "User not found")
		return
	}

	var user User
	decodeForm(w, req, &user)

	foundUser.Email = user.Email
	foundUser.Password = user.Password
	db.updateUser(foundUser)

	respondWithJSON(w, 200, foundUser.userToPublic())
}

func handlerLogin(w http.ResponseWriter, req *http.Request) {
	db, _ := createDB(DBPATH)
	var user User
	decodeForm(w, req, &user)
	foundUser, found := db.getUsersMap()[user.Email]
	if !found {
		respondWithError(w, 404, "user not found")
		return
	}
	foundUser.ExpiresInSeconds = user.ExpiresInSeconds
	err := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(user.Password))
	if err != nil {
		respondWithError(w, 401, "wrong password")
		return
	}
	db.updateUser(foundUser)
	respondWithJSON(w, 200, foundUser.UserLoginResponse())
}

func handlerAddChirp(w http.ResponseWriter, req *http.Request) {
	db, err := createDB(DBPATH)
	if err != nil {
		panic(err)
	}
	var chirp Chirp
	decodeForm(w, req, &chirp)
	chirp.Body = censor(chirp.Body)

	if len(chirp.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	chirp.ID = CHIRPID
	db.addChirp(chirp)
	respondWithJSON(w, 201, chirp)
}

func handlerGetChirps(w http.ResponseWriter, req *http.Request) {
	db, err := createDB(DBPATH)
	if err != nil {
		panic(err)
	}
	respondWithJSON(w, 200, db.getChirps())
}

func handlerAddChirpId(w http.ResponseWriter, req *http.Request) {
	chirpID, ok := mux.Vars(req)["chirpID"]
	if !ok {
		respondWithError(w, 400, "id is missing in parameters")
		return
	}
	id, err := strconv.Atoi(chirpID)
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}
	db, err := createDB(DBPATH)
	if err != nil {
		panic(err)
	}
	chirpMap, err := db.loadDB()
	if err != nil {
		panic(err)
	}
	chirp, found := chirpMap.Chirps[id]
	if !found {
		respondWithError(w, 404, "Chirp not found")
		return
	}
	respondWithJSON(w, 200, chirp)
}

func handlerAddUser(w http.ResponseWriter, req *http.Request) {
	db, err := createDB(DBPATH)
	if err != nil {
		panic(err)
	}
	var user User
	decodeForm(w, req, &user)
	_, found := db.getUsersMap()[user.Email]
	if found {
		respondWithError(w, 409, "user already exists")
		return
	}
	user.ID = USERID
	db.addUser(user)
	respondWithJSON(w, 201, user.userToPublic())
}

func censor(s string) string {
	re := regexp.MustCompile(`(?i)kerfuffle|sharbert|fornax`)
	return re.ReplaceAllString(s, "****")
}

// decodes json into your provided struct. Using this to avoid making a massive all encompassing struct
func decodeForm(w http.ResponseWriter, r *http.Request, dst interface{}) {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		respondWithError(w, 400, "unable to decode email form")
	}
}

type Login struct {
	Password         string `json:"password"`
	Email            string `json:"email"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

func (u *User) generateClaims() *jwt.RegisteredClaims {
	// 24h
	expires := time.Now().UTC().Add(time.Second * time.Duration(86400))
	claims := &jwt.RegisteredClaims{
		Issuer:    "Chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(expires),
		Subject:   u.Email,
	}
	if u.ExpiresInSeconds > 0 && u.ExpiresInSeconds < 86400 {
		claims.ExpiresAt = jwt.NewNumericDate(time.Now().UTC().Add(time.Second * time.Duration(u.ExpiresInSeconds)))
	}
	return claims
}

func (u *User) generateToken() string {
	jwtSecret := os.Getenv("JWT_SECRET")
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, u.generateClaims())
	token, err := t.SignedString([]byte(jwtSecret))
	if err != nil {
		panic(err)
	}
	return token
}

type UserTokenResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Token string `json:"token"`
}
