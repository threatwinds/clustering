package clustering

import (
	"context"
	"fmt"
	"time"

	"github.com/threatwinds/clustering/helpers"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Join is a method that handles the registration of a new node in the cluster.
// It receives the node properties and updates the cluster accordingly.
func (cluster *cluster) Join(ctx context.Context, in *NodeProperties) (*emptypb.Empty, error) {
	helpers.Logger.LogF(
		200,
		"received register request from %s",
		in.NodeIp,
	)

	newNode := cluster.newNode(in)

	if in.Timestamp > newNode.properties.Timestamp {
		newNode.properties = in
	}

	return nil, nil
}

// UpdateNode is a method that updates the properties of an existing node in the cluster.
// It receives the updated node properties and updates the cluster accordingly.
func (cluster *cluster) UpdateNode(ctx context.Context, in *NodeProperties) (*emptypb.Empty, error) {
	node, e := cluster.getNode(in.NodeIp)
	if e != nil {
		if e.Is("not found") {
			cluster.newNode(in)
			return nil, nil
		}

		return nil, fmt.Errorf(e.Message)
	}

	if in.Timestamp > node.properties.Timestamp {
		node.properties = in
	}

	return nil, nil
}

// Echo is a method that handles the ping-pong mechanism between nodes in the cluster.
// It receives a ping message from a node, updates the node's health status, and sends back a pong message.
func (cluster *cluster) Echo(ctx context.Context, in *Ping) (*Pong, error) {
	node, e := cluster.getNode(in.NodeIp)
	if e != nil {
		return nil, fmt.Errorf(e.Message)
	}

	now := time.Now().UTC().UnixMilli()

	node.setHealthy(now, in.Timestamp)

	return &Pong{
		PingTimestamp: in.Timestamp,
		PongTimestamp: now,
	}, nil
}

// ProcessTask is a method that processes incoming tasks from the client.
// It receives a gRPC stream and continuously listens for incoming tasks.
// Depending on the task's function name, it performs the corresponding action.
func (cluster *cluster) ProcessTask(srv Cluster_ProcessTaskServer) error {
	for {
		task, err := srv.Recv()
		if err != nil {
			return err
		}

		for fName, f := range cluster.callBackDict {
			if fName == task.FunctionName {
				go f(task)
			}
		}
	}
}
