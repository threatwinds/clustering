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

// node represents a node in the cluster.
type node struct {
	properties *NodeProperties
	connection *grpc.ClientConn
	lastPing   int64
	latency    int64
	tasks      chan *Task
	mutex      chan struct{}
}

// client returns a ClusterClient for the node.
func (node *node) client() ClusterClient {
	node.connect()

	return NewClusterClient(node.connection)
}

// connect establishes a connection to the node.
func (node *node) connect() *logger.Error {
	if node.connection != nil {
		return nil
	}

	helpers.Logger.LogF(200, "connecting to %s", node.properties.NodeIp)

	err := helpers.Logger.Retry(func() error {
		conn, err := grpc.Dial(
			fmt.Sprintf("%s:%d", node.properties.NodeIp,
				helpers.GetCfg().ClusterPort),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}
		
		node.connection = conn

		go node.startSending()
		
		return nil
	})
	if err != nil {
		node.setUnhealthy(err.Error())
		return helpers.Logger.ErrorF("error establishing connection: %v", err)
	}

	return nil
}

// startSending starts sending tasks to the node.
func (node *node) startSending() {
	client := NewClusterClient(node.connection)
	pTaskClient, err := client.ProcessTask(context.Background())
	if err != nil {
		helpers.Logger.ErrorF("error sending task: %v", err)
		node.setUnhealthy(err.Error())
		return
	}
	for {
		task := <-node.tasks
		err := helpers.Logger.Retry(func() error {
			err := pTaskClient.Send(task)
			return err
		})

		if err != nil {
			helpers.Logger.ErrorF("error sending task: %v", err)
			node.setUnhealthy(err.Error())
			return
		}
	}
}

// joinTo joins the current node to a new node in the cluster.
func (node *node) joinTo(newNode *node) *logger.Error {
	err := helpers.Logger.Retry(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, err := newNode.client().Join(ctx, node.properties)
		return err
	})

	if err != nil {
		node.setUnhealthy(err.Error())
		return helpers.Logger.ErrorF("error registering from %s to %s, %v", node.properties.NodeIp, newNode.properties.NodeIp, err)
	}

	return nil
}

// updateTo updates the current node's properties to the destination node.
func (node *node) updateTo(dstNode *node) *logger.Error {
	err := helpers.Logger.Retry(func() error {
		_, err := dstNode.client().UpdateNode(context.Background(), node.properties)
		return err
	})
	if err != nil {
		dstNode.setUnhealthy(err.Error())
		return helpers.Logger.ErrorF("cannot send resources update to %s: %v", dstNode.properties.NodeIp, err)
	}

	return nil
}

// withLock acquires a lock on the node and performs the specified action.
func (node *node) withLock(ref string, action func() error) *logger.Error {
	select {
	case <-time.After(10 * time.Second):
		return helpers.Logger.ErrorF("%s: timeout waiting to lock %s", ref, node.properties.NodeIp)
	case node.mutex <- struct{}{}:
		defer func() { <-node.mutex }()
	}

	err := action()
	if err != nil {
		return helpers.Logger.ErrorF("error in action: %v", err)
	}

	return nil
}
