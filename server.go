package clustering

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/threatwinds/clustering/helpers"
	"google.golang.org/protobuf/types/known/emptypb"
)

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

func (cluster *cluster) ProcessTask(srv Cluster_ProcessTaskServer) error {
	for {
		task, err := srv.Recv()
		if err != nil {
			return err
		}

		switch task.FunctionName {
		case "broadcast":
			cluster.BroadcastTask(task)
		case "enqueue":
			nodes, err := strconv.Atoi(task.Args[0])
			if err != nil {
				return err
			}
			cluster.EnqueueTask(task, nodes)
		}

		for fName, f := range cluster.callBackDict {
			if fName == task.FunctionName {
				go f(task)
			}
		}
	}
}
