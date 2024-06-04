package main

import (
	"fmt"
	"net/http"

	"github.com/pat955/chirpy/internal/my_db"
)

type Event struct {
	Event string `json:"event"`
	Data  struct {
		UserID int `json:"user_id"`
	} `json:"data"`
}

func handlerUpgraded(w http.ResponseWriter, req *http.Request) {
	var event Event
	my_db.DecodeForm(req, event)
	fmt.Println(event)
}
