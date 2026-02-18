package main

import (
	"fmt"
	"time"
	"sync"
	"log"
	"net/http"
	"encoding/json"
	"github.com/google/uuid"
)

type Message struct {
	Room    string `json:"room"`
	User    string `json:"user"`
	Message string `json:"message"`
}

type Room struct {
	mu      sync.Mutex
	clients map[chan Message]bool
}

var mu     sync.Mutex
var signal chan Message
var rooms  map[string]Room

func main() {
	go broadcast()

	signal = make(chan Message)
	rooms = make(map[string]Room)

	http.HandleFunc("GET  /api/rooms", API_GetRooms)                       // Retrieve a list of active rooms
	http.HandleFunc("GET  /api/rooms/{name}/stream", API_GetRoomStream)    // Stream messages sent in room
	http.HandleFunc("POST /api/rooms", API_PostRoom)                       // Create a room with a random name
	http.HandleFunc("POST /api/rooms/{name}", API_PostRoomNamed)           // Create a room with name
	http.HandleFunc("POST /api/rooms/{name}/message", API_PostRoomMessage) // Send message to room

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func broadcast() {
	var message Message
	var room    Room

	for {
		message = <- signal
		room = rooms[message.Room]
		room.mu.Lock()
		for client := range room.clients {
			client <- message
		}
		room.mu.Unlock()
	}
}

func API_GetRooms(w http.ResponseWriter, req *http.Request) {
	var names []string

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	for name, _ := range rooms {
		names = append(names, name)
	}

	json.NewEncoder(w).Encode(names)
}

func API_GetRoomStream(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	name := req.PathValue("name")
	room, ok := rooms[name]

	if !ok {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "stream not supported", http.StatusInternalServerError)
		return
	}

	client := make(chan Message)

	room.mu.Lock()
	room.clients[client] = true
	room.mu.Unlock()

	defer func() {
		room.mu.Lock()
		delete(room.clients, client)
		close(client)
		room.mu.Unlock()
		fmt.Println("client disconnected")
	}()

	ctx := req.Context()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-client:
			fmt.Fprintf(w, "data: [%s] %s: %s\n\n",
				time.Now().Format("15:04:05"),
				msg.User,
				msg.Message,
			)
			flusher.Flush()
		}
	}
}

func API_PostRoom(w http.ResponseWriter, req *http.Request) {
	var name string

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	name = uuid.New().String()

	rooms[name] = Room {
		clients : make(map[chan Message]bool),
	}

	json.NewEncoder(w).Encode(name)
}

func API_PostRoomNamed(w http.ResponseWriter, req *http.Request) {
	var name string

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	name = req.PathValue("name")

	rooms[name] = Room {
		clients : make(map[chan Message]bool),
	}

	json.NewEncoder(w).Encode(name)
}

func API_PostRoomMessage(w http.ResponseWriter, req *http.Request) {
	var message Message

	w.WriteHeader(http.StatusOK)

	_ = json.NewDecoder(req.Body).Decode(&message)
	message.Room = req.PathValue("name")

	signal <- message
}
