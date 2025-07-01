package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/liuisaac/choo-choo/api"
	raftnode "github.com/liuisaac/choo-choo/internal/raft"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file:", err)
	}

	nodeID := os.Getenv("NODE_ID")
	bindAddr := os.Getenv("RAFT_BIND")
	httpAddr := os.Getenv("HTTP_ADDR")
	dataDir := os.Getenv("DATA_DIR")
	peers := []string{}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("failed to create data dir: %v", err)
	}

	node, err := raftnode.NewNode(dataDir, nodeID, bindAddr, peers)
	if err != nil {
		log.Fatalf("raft node init failed: %v", err)
	}

	log.Printf("Node %s started at %s", nodeID, bindAddr)

	server := api.New(node)
	log.Fatal(server.Start(httpAddr))
}
