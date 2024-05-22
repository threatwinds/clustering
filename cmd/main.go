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

func printMsg(task *clustering.Task) {
	helpers.Logger.LogF(200, "received task %s with args %v", task.FunctionName, task.Args)
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

	e := cluster.Start(map[string]func(task *clustering.Task){
		"print": printMsg,
	})

	if e != nil {
		os.Exit(1)
	}

	for {
		nu := rand.Intn(1000)

		helpers.Logger.LogF(200, "sending %d", nu)

		cluster.BroadcastTask(&clustering.Task{
			AnswerTo:     cluster.MyIp(),
			FunctionName: "print",
			Args:         []string{fmt.Sprintf("%d", nu), cluster.MyIp()},
		})

		time.Sleep(5 * time.Second)
	}
}
