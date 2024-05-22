package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/threatwinds/clustering" 
	"github.com/threatwinds/clustering/helpers"
	"google.golang.org/grpc"
)

func printMsg(task *clustering.Task) error {
	helpers.Logger.LogF(200, "received task with args %v", task.Args)
	
	fmt.Printf("received task with args %v\n", task.Args)
	
	return nil
}

func main() {
	cluster := clustering.New()
	
	grpcServer := grpc.NewServer()

	clustering.RegisterClusterServer(grpcServer, cluster)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", helpers.GetCfg().ClusterPort))
	if err != nil {
		helpers.Logger.ErrorF(err.Error())
		os.Exit(1)
	}

	go grpcServer.Serve(lis)

	e := cluster.Start(map[string]func(task *clustering.Task) error{
		"print": printMsg,
	})

	if e != nil {
		os.Exit(1)
	}

	for {
		nu := rand.Intn(1000)

		helpers.Logger.LogF(200, "sending %d", nu)
		
		cluster.EnqueueTask(&clustering.Task{
			AnswerTo: cluster.LocalNode.Properties.NodeIp,
			FunctionName: "print",
			Args: []string{fmt.Sprintf("%d", nu)},
		}, 1)

		time.Sleep(5 * time.Second)
	}
}