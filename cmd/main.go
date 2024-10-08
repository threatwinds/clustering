package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/threatwinds/clustering"
	go_sdk "github.com/threatwinds/go-sdk"
	"google.golang.org/grpc"
)

func printMsg(task *clustering.Task) {
	go_sdk.Logger().LogF(200, "received task %s with args %v", task.FunctionName, task.Args)
}

func main() {
	clusterPort := 1993
	cluster := clustering.New(clustering.Config{
		ClusterPort: clusterPort,
		SeedNodes:   []string{"172.17.0.2"},
		DataCenter:  1,
		Rack:        1,
		LogLevel:    200,
	}, map[string]func(task *clustering.Task){
		"print": printMsg,
	})

	grpcServer := grpc.NewServer()

	clustering.RegisterClusterServer(grpcServer, cluster)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", clusterPort))
	if err != nil {
		go_sdk.Logger().ErrorF(err.Error())
		os.Exit(1)
	}

	go grpcServer.Serve(lis)

	e := cluster.Start()

	if e != nil {
		os.Exit(1)
	}

	for {
		nu := rand.Intn(1000)

		go_sdk.Logger().LogF(100, "sending %d", nu)

		cluster.BroadcastTask(&clustering.Task{
			AnswerTo:     cluster.MyIp(),
			FunctionName: "print",
			Args:         []string{fmt.Sprintf("%d", nu), cluster.MyIp()},
		})

		time.Sleep(50 * time.Second)
	}
}
