package raft

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
	"github.com/liuisaac/choo-choo/internal/kv"
)

type RaftNode struct {
	Raft  *raft.Raft
	Store *kv.Store
}

func NewNode(dataDir string, nodeID string, bindAddr string, peers []string) (*RaftNode, error) {
	store := kv.NewStore()

	// configure raft
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(nodeID)

	// create boltdb instance for WAL and stable storage
	dbPath := filepath.Join(dataDir, "raft.db")
	boltStore := must(raftboltdb.NewBoltStore(dbPath))

	// create snapshot store
	snapshotStore := must(raft.NewFileSnapshotStore(dataDir, 1, os.Stderr))

	// create TCP-based transport for communication
	transport := must(raft.NewTCPTransport(bindAddr, nil, 3, 10*time.Second, os.Stderr))

	// initialize raft instance
	raftNode := must(raft.NewRaft(config, store, boltStore, boltStore, snapshotStore, transport))

	// bootstrap cluster if first node
	if len(peers) == 0 {
		// check if this is the first node by getting the current configuration
		future := raftNode.GetConfiguration()
		if err := future.Error(); err != nil {
			return nil, fmt.Errorf("failed to get raft config: %w", err)
		}

		// only bootstrap if no servers are configured yet
		if len(future.Configuration().Servers) == 0 {
			cfg := raft.Configuration{
				Servers: []raft.Server{{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				}},
			}
			if err := raftNode.BootstrapCluster(cfg).Error(); err != nil {
				return nil, fmt.Errorf("failed to bootstrap cluster: %w", err)
			}
		}
	}

	return &RaftNode{
		Raft:  raftNode,
		Store: store,
	}, nil
}

// helper to idiomatically wrap resource-error boilerplate
func must[T any](value T, err error) T {
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}
	return value
}
