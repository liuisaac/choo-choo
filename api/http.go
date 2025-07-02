package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin" // using gin for HTTP server
	db "github.com/liuisaac/choo-choo/internal/kv"
	node "github.com/liuisaac/choo-choo/internal/raft"
	"github.com/liuisaac/choo-choo/internal/parser"
)

// QueryRequest represents the JSON structure for incoming query requests
// it expects a single field "q" which is the query string (ie: {"q": "SET key value"})
type QueryRequest struct {
	Q string `json:"q"`
}

type HTTPServer struct {
	node *node.RaftNode
}

func New(node *node.RaftNode) *HTTPServer {
	return &HTTPServer{node: node}
}

func (s *HTTPServer) Start(addr string) error {
    r := gin.Default()

    r.POST("/query", s.handleQuery)

    return r.Run(addr) // keeps the server running at addr (e.g. ":8080")
}

// handleQuery handles incoming HTTP requests for the /query endpoint
// ie: a POST request with a JSON body like {"q": "SET key value"}
// will apply the command to the Raft log and return a response
func (s *HTTPServer) handleQuery(c *gin.Context) {
	var req QueryRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	query, err := parser.ParseQuery(req.Q)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// dispatch to Raft or KVS
	switch query.Op {
	case "set", "delete":
		cmd := db.Command{
			Op:    query.Op,
			Key:   query.Key,
			Value: query.Value,
		}
		data, _ := json.Marshal(cmd)
		f := s.node.Raft.Apply(data, 5*time.Second)
		if err := f.Error(); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "ok"})
	case "get":
		value, ok := s.node.Store.Get(query.Key)
		if !ok {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}
		c.JSON(200, gin.H{"value": value})
	case "info":
		c.JSON(200, gin.H{
			"id":     s.node.Raft.String(),
			"leader": s.node.Raft.Leader(),
			"state":  s.node.Raft.State().String(),
		})
	default:
		c.JSON(400, gin.H{"error": "unsupported op"})
	}
}