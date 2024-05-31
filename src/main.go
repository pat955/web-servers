package main

import (
	"encoding/json"
	"flag"
	"os"
	"time"

	"net/http"
	"regexp"
	"strconv"

	"github.com/golang-jwt/jwt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

const DBPATH string = "./database.json"

func main() {
	godotenv.Load("../.env")
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

	claims := &jwt.RegisteredClaims{
		Issuer: "Chirpy",
		IssuedAt: time.Now(),
		ExpiresAt: jwt.NewNumericDate(time.Now()+time.Unix(120)),
		Subject: string(user.id)
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(mySigningKey)
	fmt.Println(ss, err)
	// token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims, token.SignedString)


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

func handlerAuth(w http.ResponseWriter, req *http.Request){
	bearer := req.Header.Get("Bearer")
	respondWithJSON(w, 200, UserTokenResponse{ID: user.id, })
}

func handlerLogin(w http.ResponseWriter, req *http.Request) {
	db, _ := createDB(DBPATH)

	var login Login
	decodeForm(w, req, &login)
	foundUser, found := db.getUsersMap()[login.Email]
	if !found {
		respondWithError(w, 404, "user not found")
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(login.Password))
	if err != nil {
		respondWithError(w, 401, "wrong password")
		return
	}
	respondWithJSON(w, 200, foundUser.userToPublic())

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
	respondWithJSON(w, 201, PublicUser{ID: user.ID, Email: user.Email})
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

func (u *User) generateClaims() *jwt.RegisteredClaims{
	// 24h
	expires := time.Now() + time.Unix(86400)
	claims := &jwt.RegisteredClaims{
		Issuer: "Chirpy",
		IssuedAt: time.Now(),
		ExpiresAt: jwt.NewNumericDate(expires),
		Subject: string(u.ID)
	}
	if user.ExpiresInSeconds > 0 && user.expires_in_seconds < 86400{
		claims.ExpiresAt := time.Now() + time.Unix(user.expires_in_seconds)
	}
	return claims
}
func (u *User) generateToken() {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, u.generateClaims())
	tokenString, err := token.SignedString(hmacSampleSecret)
	fmt.Println(tokenString, err)
}
type UserTokenResponse struct{
	ID int `json:"id"`
	Email string `json:"email"`
	Token string `json:"token"`
}