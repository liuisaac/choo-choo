package api

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/raft"
	raftnode "github.com/liuisaac/choo-choo/internal/raft"
)

type HTTPServer struct {
	Node *raftnode.RaftNode
}

func New(node *raftnode.RaftNode) *HTTPServer {
	return &HTTPServer{Node: node}
}

func (s *HTTPServer) Start(addr string) error {
	http.HandleFunc("/get", s.get)
	http.HandleFunc("/set", s.set)
	http.HandleFunc("/delete", s.delete)
	http.HandleFunc("/join", s.join)
	return http.ListenAndServe(addr, nil)
}

func (s *HTTPServer) get(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	val, ok := s.Node.Store.Get(key)
	if !ok {
		http.NotFound(w, r)
		return
	}
	io.WriteString(w, val)
}

func (s *HTTPServer) set(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	cmd := map[string]string{
		"op":    "set",
		"key":   req.Key,
		"value": req.Value,
	}
	data, _ := json.Marshal(cmd)
	f := s.Node.Raft.Apply(data, 5*time.Second)
	if err := f.Error(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *HTTPServer) delete(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	cmd := map[string]string{
		"op":  "delete",
		"key": key,
	}
	data, _ := json.Marshal(cmd)
	f := s.Node.Raft.Apply(data, 5*time.Second)
	if err := f.Error(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *HTTPServer) join(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID   string `json:"id"`
		Addr string `json:"addr"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid join request", http.StatusBadRequest)
		return
	}

	err := s.Node.Raft.AddVoter(raft.ServerID(req.ID), raft.ServerAddress(req.Addr), 0, 0).Error()
	if err != nil {
		http.Error(w, "failed to add voter: "+err.Error(), http.StatusInternalServerError)
	}
}