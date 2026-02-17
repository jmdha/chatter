package main

import (
	"sync"
	"fmt"
	"time"
	"log"
	"encoding/json"
	"net/http"
)

type Message struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

var (
	mu sync.Mutex
	clients   = make(map[chan Message]bool)
	broadcast = make(chan Message)
)

func main() {
	go handleBroadcast()

	http.HandleFunc("GET /", RouteGetStream)
	http.HandleFunc("POST /", RoutePostStream)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleBroadcast() {
	for {
		msg := <-broadcast

		mu.Lock()
		for client := range clients {
			select {
			case client <- msg:
			default:
			}
		}
		mu.Unlock()
	}
}

func RoutePostStream(w http.ResponseWriter, req *http.Request) {
	var message Message

	if err := json.NewDecoder(req.Body).Decode(&message); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if message.User == "" || message.Message == "" {
		http.Error(w, "requires user and message", http.StatusBadRequest)
		return
	}

	broadcast <- message

	w.WriteHeader(http.StatusOK)
}

func RouteGetStream(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "stream not supported", http.StatusInternalServerError)
		return
	}

	client := make(chan Message)

	mu.Lock()
	clients[client] = true
	mu.Unlock()

	defer func() {
		mu.Lock()
		delete(clients, client)
		close(client)
		mu.Unlock()
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
