package kv

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/hashicorp/raft"
)

// state machine for the kv store

// Command is a struct that represents a mutation to the kv store
// note: currently only supports set and delete operations
type Command struct {
	Op string `json:"op"`
	Key string `json:"key"`
	Value string `json:"value,omitempty"`
}

// Store is a Raft-backed thread-safe in-memory kv store
//
// in terms of raft, this is the the state machine applied to the log
// it should not be directly exposed to clients except through controlled APIs
//
// all state transitions occur exclusively through Raft log application via Apply()
// Reads are permitted via lock-protected access. No external mutation is allowed
//
// invariants
// - Apply() must be deterministic and side-effect-free except for store mutation
// - store accesses must be guarded by rw mutex to handle concurrency
// - store ONLY REFLECTS committed log state
type Store struct {
	mu sync.RWMutex				// mutex enforced lock to prioritize strong consistency over throughput
	store map[string]string		// in-memory map to hold key-value pairs (nondurable)
}

// NewStore initializes a new Store instance and returns a pointer to that instance
func NewStore() *Store {
	return &Store{
		store: make(map[string]string),
	}
}

// Apply commits a Raft log entry to the store
func (s *Store) Apply(l *raft.Log) interface{} {
	var cmd Command
	if err := json.Unmarshal(l.Data, &cmd); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	switch cmd.Op {
	case "set":
		s.store[cmd.Key] = cmd.Value
	case "delete":
		delete(s.store, cmd.Key)
	default:
		return io.ErrUnexpectedEOF
	}

	return nil
}

// Get reads a value from the store by key
func (s *Store) Get (key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.store[key]
	return value, exists
}

// Snapshot implements the raft.FSMSnapshot inferface
//
// represents a snapshot (immutable) of the kv store state
// this is captured during log compaction in Raft, used internally
// by Raft to persist state to disk
//
// lifecycle
// - created by Store.Snapshot() during snapshotting ops
// - persisted to disk by Persist() method
// - released by Release() method
//
// invariants
// - Must represent a consistent, immutable view of the state at snapshot time
// - Must not perform any mutations
// - Must fully serialize itself via Persist()
type Snapshot struct {
	state map[string]string
}

// Snapshot creates a snapshot of the current state of the store
// note: this should only be called during snapshotting ops and while the db is in a consistent state
func (s *Store) Snapshot() (raft.FSMSnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// generate deep copy of store to freeze db state
	state := make(map[string]string, len(s.store))
	for k, v := range s.store {
		state[k] = v
	}

	return &Snapshot{state: state}, nil
}

// Persist writes the snapshot to the given sink
func (s *Snapshot) Persist(sink raft.SnapshotSink) error {
	data, err := json.Marshal(s.state)
	if err != nil {
		sink.Cancel()
		return err
	}
	if _, err := sink.Write(data); err != nil {
		sink.Cancel()
		return err
	}
	return sink.Close()
}

// Release is a no-op for this in-memory snapshot
// note: since this is an in-memory snapshot, we don't need to do anything here
func (s *Snapshot) Release() {}

// Restore replaces the store's state with a snapshot from disk
// this is called by Raft when a node needs to catch up quickly (ie: node restart or when joining the cluster)
func (s *Store) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	s.mu.Lock()
	defer s.mu.Unlock()

	var store map[string]string
	if err := json.NewDecoder(rc).Decode(&store); err != nil {
		return err
	}

	s.store = store
	return nil
}