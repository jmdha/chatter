package main

import (
	"fmt"
	"chatter/api"
	"encoding/json"
	"net/http"
)

func (h *Handler) GET_Index(w http.ResponseWriter, req *http.Request) {

}

func (h *Handler) GET_API(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "stream not supported", http.StatusInternalServerError)
		return
	}

	line_name := req.PathValue("name")
	channel := make(chan api.Message)
	line := h.line(line_name)
	line.Subscribe(channel)
	defer line.Unsubscribe(channel)

	ctx := req.Context()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-channel:
			fmt.Fprintf(w, "data: [%s] <%s>: %s\n\n",
				msg.Timestamp,
				msg.User,
				msg.Message,
			)
			flusher.Flush()
		}
	}
}

func (h *Handler) POST_API(w http.ResponseWriter, req *http.Request) {
	line_name := req.PathValue("name")

	line, ok := h.lines[line_name]
	if !ok {
		return
	}

	var msg api.Message
	err := json.NewDecoder(req.Body).Decode(&msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	line.Publish(msg)

	w.WriteHeader(http.StatusNoContent)
}
