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
	"github.com/threatwinds/logger"
)

// cluster represents a cluster of nodes.
type cluster struct {
	localNode    *node
	nodes        map[string]*node
	callBackDict map[string]func(task *Task)
	mutex        chan struct{}
	UnimplementedClusterServer
}

var clusterInstance *cluster
var clusterOnce sync.Once

// New creates a new instance of the cluster.
func New() *cluster {
	clusterOnce.Do(func() {
		clusterInstance = &cluster{}
		clusterInstance.nodes = make(map[string]*node, 3)
	})

	return clusterInstance
}

// withLock acquires a lock on the cluster and performs the specified action.
func (cluster *cluster) withLock(ref string, action func() error) *logger.Error {
	wait, _ := time.ParseDuration(fmt.Sprintf("%ds", len(cluster.nodes)*2))

	if wait == 0 {
		wait = 2 * time.Second
	}

	select {
	case <-time.After(wait):
		return helpers.Logger.ErrorF("%s: timeout waiting to lock cluster", ref)
	case cluster.mutex <- struct{}{}:
		defer func() { <-cluster.mutex }()
	}

	err := action()
	if err != nil {
		return helpers.Logger.ErrorF("error in action: %v", err)
	}

	return nil
}

// Start starts the cluster with the specified callback dictionary.
func (cluster *cluster) Start(callBackDict map[string]func(task *Task)) *logger.Error {
	helpers.Logger.LogF(200, "starting cluster")

	var e *logger.Error

	cluster.callBackDict = callBackDict

	cluster.mutex = make(chan struct{}, 1)

	cluster.localNode = new(node)

	cluster.localNode.latency = -1

	cluster.localNode.tasks = make(chan *Task)

	cluster.localNode.mutex = make(chan struct{}, 1)

	cluster.localNode.properties = new(NodeProperties)

	cluster.localNode.properties.NodeIp, e = helpers.GetMainIP()
	if e != nil {
		return e
	}

	cluster.localNode.properties.UpSince = time.Now().UTC().UnixMilli()

	cluster.localNode.properties.Status = "new"

	cluster.localNode.properties.DataCenter = helpers.GetCfg().DataCenter

	cluster.localNode.properties.Rack = helpers.GetCfg().Rack

	cluster.nodes[cluster.localNode.properties.NodeIp] = cluster.localNode

	cluster.connectToSeeds()

	time.Sleep(15 * time.Second)

	go cluster.updateResources()

	go cluster.echo()

	time.Sleep(5 * time.Second)

	go cluster.checkNodes()

	go cluster.viralizeStatus()

	return nil
}

// connectToSeeds connects to the seed nodes in the cluster.
func (cluster *cluster) connectToSeeds() {
	cluster.withLock("connect to seed", func() error {
		for _, seed := range helpers.GetCfg().SeedNodes {
			cluster.newEmptyNode(seed)
		}

		return nil
	})
}

// healthyNodes returns a list of healthy nodes in the cluster.
func (cluster *cluster) healthyNodes() []*node {
	var nodes = make([]*node, 0, 3)

	for _, node := range cluster.nodes {
		if node.properties.Status == "unhealthy" {
			continue
		}

		nodes = append(nodes, node)
	}

	return nodes
}

// MyIp returns the IP address of the local node.
func (cluster *cluster) MyIp() string {
	return cluster.localNode.properties.NodeIp
}

// getNode returns the node with the specified name.
func (cluster *cluster) getNode(name string) (*node, *logger.Error) {
	node, ok := cluster.nodes[name]
	if !ok {
		return nil, helpers.Logger.ErrorF("node not found")
	}

	return node, nil
}

// getRandomNode returns a random healthy node from the cluster. It excludes the local node.
func (cluster *cluster) getRandomNode() (*node, *logger.Error) {
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
func (cluster *cluster) newEmptyNode(ip string) *node {
	var newNode *node

	for _, node := range cluster.nodes {
		if node.properties.NodeIp == ip {
			newNode = node
		}
	}

	if newNode == nil {
		newNode = &node{
			properties: &NodeProperties{
				NodeIp: ip,
				Status: "new",
			},
			tasks: make(chan *Task),
			mutex: make(chan struct{}, 1),
		}

		cluster.nodes[newNode.properties.NodeIp] = newNode

		cluster.localNode.joinTo(newNode)
	}

	return newNode
}

// newNode creates a new node with the specified properties if it does not exist and joins it to the cluster.
func (cluster *cluster) newNode(properties *NodeProperties) *node {
	var newNode *node

	for _, node := range cluster.nodes {
		if node.properties.NodeIp == properties.NodeIp {
			newNode = node
		}
	}

	if newNode == nil {
		newNode = &node{
			properties: properties,
			tasks:      make(chan *Task),
			mutex:      make(chan struct{}, 1),
		}

		cluster.nodes[newNode.properties.NodeIp] = newNode

		cluster.localNode.joinTo(newNode)
	}

	return newNode
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

			cluster.withLock("sending resources update", func() error {
				for _, node := range cluster.healthyNodes() {
					if node.properties.NodeIp == cluster.localNode.properties.NodeIp {
						continue
					}

					node.withLock("sending resources update", func() error {
						cluster.localNode.updateTo(node)

						return nil
					})
				}

				return nil
			})

			return nil
		})

		time.Sleep(10 * time.Second)
	}
}

// viralizeStatus sends the status updates of the local node to other nodes in the cluster.
func (cluster *cluster) viralizeStatus() {
	for {
		cluster.withLock("sending resources update", func() error {
			for _, node := range cluster.nodes {
				if node.properties.NodeIp == cluster.localNode.properties.NodeIp {
					continue
				}

				alreadySent := make(map[string]bool)

				for len(alreadySent) < len(cluster.healthyNodes()) && len(alreadySent) < 3 {
					rNode, e := cluster.getRandomNode()
					if e != nil {
						continue
					}

					if _, ok := alreadySent[rNode.properties.NodeIp]; !ok {
						rNode.withLock("viralizing node resources", func() error {
							node.updateTo(rNode)
							alreadySent[rNode.properties.NodeIp] = true

							return nil
						})
					}
				}
			}

			return nil
		})

		time.Sleep(10 * time.Second)
	}
}

// echo sends a ping to other nodes in the cluster and checks their response.
func (cluster *cluster) echo() {
	for {
		cluster.withLock("sending echo", func() error {
			for _, node := range cluster.nodes {
				ping := &Ping{
					Timestamp: time.Now().UTC().UnixMilli(),
					NodeIp:    cluster.localNode.properties.NodeIp,
				}

				node.withLock("sending echo", func() error {
					pong, err := node.client().Echo(context.Background(), ping)
					if err != nil {
						node.setUnhealthy(err.Error())
						return fmt.Errorf("cannot send ping to %s: %v", node.properties.NodeIp, err)
					}

					if pong.PongTimestamp-pong.PingTimestamp > 30000 {
						node.setUnhealthy("latency too high")
					}

					return nil
				})
			}

			return nil
		})

		time.Sleep(1 * time.Second)
	}
}
