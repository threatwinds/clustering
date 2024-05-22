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

type node struct {
	properties *NodeProperties
	connection *grpc.ClientConn
	lastPing   int64
	latency    int64
	tasks      chan *Task
	mutex      chan struct{}
}

func (node *node) client() ClusterClient {
	node.connect()

	return NewClusterClient(node.connection)
}

func (node *node) connect() *logger.Error {
	if node.connection != nil {
		return nil
	}

	helpers.Logger.LogF(200, "connecting to %s", node.properties.NodeIp)

	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", node.properties.NodeIp,
			helpers.GetCfg().ClusterPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return helpers.Logger.ErrorF("error establishing connection: %v", err)
	}

	node.connection = conn

	go node.startSending()

	return nil
}

func (node *node) startSending() {
	client := NewClusterClient(node.connection)
	pTaskClient, err := client.ProcessTask(context.Background())
	if err != nil {
		helpers.Logger.ErrorF("error processing task: %v", err)
		return
	}
	for {
		task := <-node.tasks
		err := pTaskClient.Send(task)
		if err != nil {
			helpers.Logger.ErrorF("error processing task: %v", err)
			node.setUnhealthy(err.Error())
			return
		}
	}
}

func (node *node) joinTo(newNode *node) *logger.Error {
	err := helpers.Logger.Retry(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, err := newNode.client().Join(ctx, node.properties)
		return err
	})

	if err != nil {
		return helpers.Logger.ErrorF("error registering from %s to %s, %v", node.properties.NodeIp, newNode.properties.NodeIp, err)
	}

	return nil
}

func (node *node) updateTo(dstNode *node) *logger.Error {
	if dstNode.properties.Status != "healthy" {
		return helpers.Logger.ErrorF("node %s is not healthy", dstNode.properties.NodeIp)
	}

	_, err := dstNode.client().UpdateNode(context.Background(), node.properties)
	if err != nil {
		e := helpers.Logger.ErrorF("cannot send resources update to %s: %v", dstNode.properties.NodeIp, err)
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

func (node *node) withLock(ref string, action func() error) *logger.Error {
	select {
	case <-time.After(5 * time.Second):
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
