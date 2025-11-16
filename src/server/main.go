package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/AdamAzuddin/snakeup/server/internal/game"
	gamemanager "github.com/AdamAzuddin/snakeup/server/internal/game_manager"
	"github.com/gorilla/mux"
)

func main() {
	fmt.Println("Starting server on :42069...")

	gameManager := &gamemanager.GameManager{Games: make(map[string]*game.Game),}
	r := mux.NewRouter()
	r.HandleFunc("/game/{gameId}", gameManager.GameHandler)
	r.HandleFunc("/api/join/{gameId}", gameManager.HandleRoomCapacity).Methods("POST", "OPTIONS")
	r.HandleFunc("/check-game/{gameId}", gameManager.CheckGameJoinable).Methods("GET", "OPTIONS")

	err := http.ListenAndServe(":42069", r)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}