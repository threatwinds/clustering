package clustering

import (
	"context"
	"fmt"
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
	mutex      chan struct{}
}

func (node *Node) client() ClusterClient {
	node.connect()

	return NewClusterClient(node.Connection)
}

func (node *Node) connect() *logger.Error {
	if node.Connection != nil {
		return nil
	}

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

func (node *Node) updateTo(dstNode *Node) *logger.Error {
	if dstNode.Properties.Status != "healthy" {
		return helpers.Logger.ErrorF("node %s is not healthy", dstNode.Properties.NodeIp)
	}

	_, err := dstNode.client().UpdateNode(context.Background(), node.Properties)
	if err != nil {
		e := helpers.Logger.ErrorF("cannot send resources update to %s: %v", dstNode.Properties.NodeIp, err)
		if e.Is("node not found") {
			node.joinTo(dstNode)
			return nil
		} else {
			dstNode.setUnhealthy(err.Error())
			return e
		}
	}

	return nil
}

func (node *Node) withLock(ref string, action func() error) *logger.Error {
	select {
	case <-time.After(5 * time.Second):
		return helpers.Logger.ErrorF("%s: timeout waiting to lock %s", ref, node.Properties.NodeIp)
	case node.mutex <- struct{}{}:
		defer func() { <-node.mutex }()
	}

	err := action()
	if err != nil {
		return helpers.Logger.ErrorF("error in action: %v", err)
	}

	return nil
}
