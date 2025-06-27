package kv

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/hashicorp/raft"
)

type Command struct {
	op string `json:"op"`
	key string `json:"key"`
	value string `json:"value,omitempty"`
}

type Store struct {
	mu sync.RWMutex
	store map[string]string
}

func NewStore() *Store {
	return &Store{
		store: make(map[string]string),
	}
}

func (s *Store) Apply(l *raft.Log) interface{} {
	var cmd Command
	if err := json.Unmarshal(l.Data, &cmd); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	switch cmd.op {
	case "set":
		s.store[cmd.key] = cmd.value
	case "delete":
		delete(s.store, cmd.key)
	default:
		return io.ErrUnexpectedEOF // or some other error
	}

	return nil
}

