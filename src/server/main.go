package main

import (
	"fmt"
	"log"
	"net/http"

	gamemanager "github.com/AdamAzuddin/snakeup/server/internal/game_manager"
	"github.com/gorilla/mux"
)

func main() {
	fmt.Println("Starting server on :42069...")

	gameManager := &gamemanager.GameManager{}
	r := mux.NewRouter()
	r.HandleFunc("/game/{gameId}", gameManager.GameHandler)

	err := http.ListenAndServe(":42069", r)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
