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
	port       int
	tasks      chan *Task
	mutex      chan struct{}
}

// client returns a ClusterClient for the node.
func (node *node) client() ClusterClient {
	node.connect()

	return NewClusterClient(node.connection)
}

// connect establishes a connection to the node.
func (node *node) connect() {
	if node.connection != nil {
		return
	}

	helpers.Logger().LogF(200, "connecting to %s", node.properties.NodeIp)

	err := helpers.Logger().Retry(func() error {
		conn, err := grpc.Dial(
			fmt.Sprintf("%s:%d", node.properties.NodeIp,
				node.port),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}

		node.connection = conn

		go node.startSending()

		return nil
	})

	if err != nil {
		e := helpers.Logger().ErrorF("error connecting to %s: %v", node.properties.NodeIp, err)
		node.setUnhealthy(e.Message)
	}
}

// startSending starts sending tasks to the node.
func (node *node) startSending() {
	client := NewClusterClient(node.connection)
	pTaskClient, err := client.ProcessTask(context.Background())
	if err != nil {
		e := helpers.Logger().ErrorF("error sending task: %v", err)
		node.setUnhealthy(e.Message)
		return
	}
	for {
		task := <-node.tasks
		err := helpers.Logger().Retry(func() error {
			err := pTaskClient.Send(task)
			return err
		})

		if err != nil {
			e := helpers.Logger().ErrorF("error sending task: %v", err)
			node.setUnhealthy(e.Message)
			return
		}
	}
}

// joinTo joins the current node to a new node in the cluster.
func (node *node) joinTo(newNode *node) {
	err := helpers.Logger().Retry(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, err := newNode.client().Join(ctx, node.properties)
		return err
	})

	if err != nil {
		e := helpers.Logger().ErrorF("error registering from %s to %s, %v", node.properties.NodeIp, newNode.properties.NodeIp, err)
		node.setUnhealthy(e.Message)
	}
}

// updateTo updates the current node's properties to the destination node.
func (node *node) updateTo(dstNode *node) {
	err := helpers.Logger().Retry(func() error {
		_, err := dstNode.client().UpdateNode(context.Background(), node.properties)
		return err
	})
	if err != nil {
		e := helpers.Logger().ErrorF("cannot send resources update to %s: %v", dstNode.properties.NodeIp, err)
		dstNode.setUnhealthy(e.Message)
	}
}

// withLock acquires a lock on the node and performs the specified action.
func (node *node) withLock(ref string, action func() error) *logger.Error {
	select {
	case <-time.After(10 * time.Second):
		return helpers.Logger().ErrorF("%s: timeout waiting to lock %s", ref, node.properties.NodeIp)
	case node.mutex <- struct{}{}:
		defer func() {
			<-node.mutex
			helpers.Logger().LogF(100, "%s: released node %s", ref, node.properties.NodeIp)
		}()
		helpers.Logger().LogF(100, "%s: locked node %s", ref, node.properties.NodeIp)
	}

	err := action()
	if err != nil {
		return helpers.Logger().ErrorF("error in action: %v", err)
	}

	return nil
}
