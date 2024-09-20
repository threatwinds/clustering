// Package clustering provides functionality for managing a cluster of nodes.
package clustering

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/threatwinds/clustering/helpers"
	go_sdk "github.com/threatwinds/go-sdk"
	"github.com/threatwinds/logger"
)

type Config struct {
	ClusterPort int
	SeedNodes   []string
	DataCenter  int32
	Rack        int32
	LogLevel    int
}

// cluster represents a cluster of nodes.
type cluster struct {
	localNode    *node
	nodes        map[string]*node
	callBackDict map[string]func(task *Task)
	mutex        chan struct{}
	config       Config
	UnimplementedClusterServer
}

var clusterInstance *cluster
var clusterOnce sync.Once

// New creates a new instance of the cluster.
func New(config Config, callBackDict map[string]func(task *Task)) *cluster {
	clusterOnce.Do(func() {
		clusterInstance = &cluster{}
		clusterInstance.nodes = make(map[string]*node, 3)
		clusterInstance.config = config
		clusterInstance.callBackDict = callBackDict
	})

	return clusterInstance
}

// withLock acquires a lock on the cluster and performs the specified action.
func (cluster *cluster) withLock(ref string, action func() error) *logger.Error {
	wait, _ := time.ParseDuration(fmt.Sprintf("%ds", len(cluster.nodes)*20))

	if wait == 0 {
		wait = 20 * time.Second
	}

	select {
	case <-time.After(wait):
		return go_sdk.Logger().ErrorF("%s: timeout waiting to lock cluster", ref)
	case cluster.mutex <- struct{}{}:
		defer func() {
			<-cluster.mutex
			go_sdk.Logger().LogF(100, "%s: released cluster", ref)
		}()
		go_sdk.Logger().LogF(100, "%s: locked cluster", ref)
	}

	err := action()
	if err != nil {
		return go_sdk.Logger().ErrorF("error in action: %v", err)
	}

	return nil
}

// Start starts the cluster with the specified callback dictionary.
func (cluster *cluster) Start() *logger.Error {
	go_sdk.Logger().LogF(200, "starting cluster")

	var e *logger.Error

	cluster.mutex = make(chan struct{}, 1)

	cluster.localNode = new(node)

	cluster.localNode.latency = -1

	cluster.localNode.tasks = make(chan *Task)

	cluster.localNode.mutex = make(chan struct{}, 1)

	cluster.localNode.port = cluster.config.ClusterPort

	cluster.localNode.properties = new(NodeProperties)

	cluster.localNode.properties.NodeIp, e = helpers.GetMainIP()
	if e != nil {
		return e
	}

	cluster.localNode.properties.UpSince = time.Now().UTC().UnixMilli()

	cluster.localNode.properties.Status = "new"

	cluster.localNode.properties.DataCenter = cluster.config.DataCenter

	cluster.localNode.properties.Rack = cluster.config.Rack

	cluster.nodes[cluster.localNode.properties.NodeIp] = cluster.localNode

	cluster.connectToSeeds()

	time.Sleep(15 * time.Second)

	go cluster.echo()

	time.Sleep(15 * time.Second)

	go cluster.updateResources()

	time.Sleep(15 * time.Second)

	go cluster.checkNodes()

	time.Sleep(15 * time.Second)

	go cluster.viralizeStatus()

	return nil
}

// connectToSeeds connects to the seed nodes in the cluster.
func (cluster *cluster) connectToSeeds() {
	for _, seed := range cluster.config.SeedNodes {
		cluster.newEmptyNode(seed)
	}
}

// healthyNodes returns a list of healthy nodes in the cluster.
func (cluster *cluster) healthyNodes() []*node {
	var nodes = make([]*node, 0, 3)

	cluster.withLock("healthy nodes", func() error {
		for _, node := range cluster.nodes {
			if node.properties.Status == "unhealthy" {
				continue
			}

			nodes = append(nodes, node)
		}

		return nil
	})

	return nodes
}

// allNodes returns a list of nodes in the cluster.
func (cluster *cluster) allNodes() []*node {
	var nodes = make([]*node, 0, 3)

	cluster.withLock("all nodes", func() error {
		for _, node := range cluster.nodes {
			nodes = append(nodes, node)
		}

		return nil
	})

	return nodes
}

// MyIp returns the IP address of the local node.
func (cluster *cluster) MyIp() string {
	return cluster.localNode.properties.NodeIp
}

// getNode returns the node with the specified name.
func (cluster *cluster) getNode(name string) (*node, *logger.Error) {
	var node *node
	var ok bool

	cluster.withLock("get node", func() error {
		node, ok = cluster.nodes[name]
		return nil
	})

	if !ok {
		return nil, go_sdk.Logger().ErrorF("node not found")
	}

	return node, nil
}

// getRandomNode returns a random healthy node from the cluster. It excludes the local node.
func (cluster *cluster) getRandomNode() (*node, *logger.Error) {
	if len(cluster.healthyNodes()) <= 1 {
		return nil, go_sdk.Logger().ErrorF("no other healthy nodes in the cluster")
	}

	for {
		n := rand.Intn(len(cluster.healthyNodes()) - 1)
		i := 0
		for _, node := range cluster.healthyNodes() {
			if i == n {
				if node.properties.NodeIp == cluster.localNode.properties.NodeIp {
					continue
				}

				return node, nil
			}

			i++
		}
	}
}

// newEmptyNode creates a new empty node with the specified IP address if it does not exist and joins it to the cluster.
func (cluster *cluster) newEmptyNode(ip string) (*node, *logger.Error) {
	var newNode *node

	for _, node := range cluster.allNodes() {
		if node.properties.NodeIp == ip {
			newNode = node
		}
	}

	e := cluster.withLock("new node", func() error {
		if newNode == nil {
			newNode = &node{
				properties: &NodeProperties{
					NodeIp: ip,
					Status: "new",
				},
				port:  cluster.config.ClusterPort,
				tasks: make(chan *Task),
				mutex: make(chan struct{}, 1),
			}

			cluster.nodes[newNode.properties.NodeIp] = newNode

			go cluster.localNode.joinTo(newNode)
		}
		return nil
	})

	return newNode, e
}

// newNode creates a new node with the specified properties if it does not exist and joins it to the cluster.
func (cluster *cluster) newNode(properties *NodeProperties) (*node, *logger.Error) {
	var newNode *node

	for _, node := range cluster.allNodes() {
		if node.properties.NodeIp == properties.NodeIp {
			newNode = node
		}
	}

	e := cluster.withLock("new node", func() error {
		if newNode == nil {
			newNode = &node{
				properties: properties,
				port:       cluster.config.ClusterPort,
				tasks:      make(chan *Task),
				mutex:      make(chan struct{}, 1),
			}

			cluster.nodes[newNode.properties.NodeIp] = newNode

			go cluster.localNode.joinTo(newNode)
		}

		return nil
	})

	return newNode, e
}

// updateResources updates the resources of the local node and sends updates to other nodes in the cluster.
func (cluster *cluster) updateResources() {
	for {
		cluster.localNode.withLock("sending local resources update", func() error {
			cpu := runtime.NumCPU()
			cluster.localNode.properties.Cores = int32(cpu)
			cluster.localNode.properties.RunningThreads = int32(runtime.NumGoroutine())

			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)
			cluster.localNode.properties.Memory = int64(mem.HeapSys)
			cluster.localNode.properties.MemoryInUse = int64(mem.HeapInuse)

			cluster.localNode.properties.Timestamp = time.Now().UTC().UnixMilli()

			for _, node := range cluster.healthyNodes() {
				if node.properties.NodeIp == cluster.localNode.properties.NodeIp {
					continue
				}

				cluster.localNode.updateTo(node)
			}

			return nil
		})

		time.Sleep(10 * time.Second)
	}
}

// viralizeStatus sends the status updates of the local node to other nodes in the cluster.
func (cluster *cluster) viralizeStatus() {
	for {
		for _, node := range cluster.healthyNodes() {
			if node.properties.NodeIp == cluster.localNode.properties.NodeIp {
				continue
			}

			alreadySent := make(map[string]bool)

			for len(alreadySent) < len(cluster.healthyNodes()) && len(alreadySent) < 3 {
				rNode, e := cluster.getRandomNode()
				if e != nil {
					time.Sleep(10 * time.Second)
					continue
				}

				if _, ok := alreadySent[rNode.properties.NodeIp]; !ok {
					node.updateTo(rNode)
					alreadySent[rNode.properties.NodeIp] = true
				}
			}
		}

		time.Sleep(10 * time.Second)
	}
}

// echo sends a ping to other nodes in the cluster and checks their response.
func (cluster *cluster) echo() {
	for {
		for _, node := range cluster.allNodes() {
			ping := &Ping{
				Timestamp: time.Now().UTC().UnixMilli(),
				NodeIp:    cluster.localNode.properties.NodeIp,
			}

			pong, err := node.client().Echo(context.Background(), ping)
			if err != nil {
				node.setUnhealthy(err.Error())
				continue
			}

			if pong.PongTimestamp-pong.PingTimestamp > 30000 {
				node.setUnhealthy("latency too high")
			}
		}

		time.Sleep(1 * time.Second)
	}
}
