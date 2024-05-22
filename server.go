package clustering

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/threatwinds/clustering/helpers"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (cluster *Cluster) Join(ctx context.Context, in *NodeProperties) (*emptypb.Empty, error) {
	helpers.Logger.LogF(
		200,
		"received register request from %s",
		in.NodeIp,
	)

	newNode := cluster.NewNode(in)

	if in.Timestamp > newNode.Properties.Timestamp {
		newNode.Properties = in
	}

	return nil, nil
}

func (cluster *Cluster) UpdateNode(ctx context.Context, in *NodeProperties) (*emptypb.Empty, error) {
	node, e := cluster.GetNode(in.NodeIp)
	if e != nil {
		return nil, fmt.Errorf(e.Message)
	}

	if in.Timestamp < node.Properties.Timestamp {
		return nil, nil
	}

	node.Properties = in

	return nil, nil
}

func (cluster *Cluster) Echo(ctx context.Context, in *Ping) (*Pong, error) {
	node, e := cluster.GetNode(in.NodeIp)
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

func (cluster *Cluster) ProcessTask(srv Cluster_ProcessTaskServer) error {
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

		for fName, f := range cluster.callBackDict{
			if fName == task.FunctionName {
				if err := f(task); err != nil {
					return err
				}
			}
		}
	}
}
