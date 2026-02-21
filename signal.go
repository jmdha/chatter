package main

import (
	"chatter/api"
	"sync"
)

type Signal struct {
	mu    sync.Mutex
	slots map[chan api.Message]bool
}

func NewSignal() Signal {
	return Signal {
		slots : make(map[chan api.Message]bool),
	}
}

func (s *Signal) Publish(msg api.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// TODO: handle slow clients
	for slot := range s.slots {
		slot <- msg
	}
}

func (s *Signal) Subscribe(c chan api.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.slots[c] = true
}

func (s *Signal) Unsubscribe(c chan api.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.slots, c)
}
