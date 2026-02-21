package main

import (
	"sync"
	"net/http"
)

type Handler struct {
	mux   *http.ServeMux
	mu    sync.Mutex
	lines map[string]Signal
}

func NewHandler() Handler {
	h := Handler {
		mux   : http.NewServeMux(),
		lines : make(map[string]Signal),
	}

	h.mux.HandleFunc("GET  /",           h.GET_Index)
	h.mux.HandleFunc("GET  /api/{name}", h.GET_API)
	h.mux.HandleFunc("POST /api/{name}", h.POST_API)

	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.mux.ServeHTTP(w, req)
}

func (h *Handler) line(name string) Signal {
	h.mu.Lock()
	defer h.mu.Unlock()
	l, ok := h.lines[name]
	if !ok {
		h.lines[name] = NewSignal()
		l, _ = h.lines[name]
	}
	return l
}
