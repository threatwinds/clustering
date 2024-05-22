package clustering

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/threatwinds/clustering/helpers"
	"github.com/threatwinds/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Node struct {
	Properties *NodeProperties
	Connection *grpc.ClientConn
	LastPing   int64
	Latency    int64
	tasks      chan *Task
	mutex      sync.RWMutex
}

func (node *Node) client() ClusterClient {
	node.connect()

	return NewClusterClient(node.Connection)
}

func (node *Node) connect() *logger.Error {
	if node.Connection != nil {
		return nil
	}

	node.mutex.Lock()
	defer node.mutex.Unlock()

	helpers.Logger.LogF(200, "connecting to %s", node.Properties.NodeIp)

	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", node.Properties.NodeIp,
			helpers.GetCfg().ClusterPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return helpers.Logger.ErrorF("error establishing connection: %v", err)
	}

	node.Connection = conn

	return nil
}

func (node *Node) joinTo(newNode *Node) *logger.Error {
	err := helpers.Logger.Retry(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, err := newNode.client().Join(ctx, node.Properties)
		return err
	})

	if err != nil {
		return helpers.Logger.ErrorF("error registering from %s to %s, %v", node.Properties.NodeIp, newNode.Properties.NodeIp, err)
	}

	return nil
}