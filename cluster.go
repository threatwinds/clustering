package clustering

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/threatwinds/clustering/helpers"
	"github.com/threatwinds/logger"
)

type Cluster struct {
	LocalNode    *Node
	Nodes        map[string]*Node
	callBackDict map[string]func(task *Task) error
	mutex        chan struct{}
	UnimplementedClusterServer
}

var clusterInstance *Cluster
var clusterOnce sync.Once

func New() *Cluster {
	clusterOnce.Do(func() {
		clusterInstance = &Cluster{}
		clusterInstance.Nodes = make(map[string]*Node, 3)
	})

	return clusterInstance
}

func (cluster *Cluster) withLock(ref string, action func() error) *logger.Error {
	select {
	case <-time.After(15 * time.Second):
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

func (cluster *Cluster) Start(callBackDict map[string]func(task *Task) error) *logger.Error {
	helpers.Logger.LogF(200, "starting cluster")

	var e *logger.Error

	cluster.callBackDict = callBackDict

	cluster.mutex = make(chan struct{}, 1)

	cluster.LocalNode = new(Node)

	cluster.LocalNode.Latency = -1

	cluster.LocalNode.tasks = make(chan *Task, 100)

	cluster.LocalNode.mutex = make(chan struct{}, 1)

	cluster.LocalNode.Properties = new(NodeProperties)

	cluster.LocalNode.Properties.NodeIp, e = helpers.GetMainIP()
	if e != nil {
		return e
	}

	cluster.LocalNode.Properties.UpSince = time.Now().UTC().UnixMilli()

	cluster.LocalNode.Properties.Status = "new"

	cluster.LocalNode.Properties.DataCenter = helpers.GetCfg().DataCenter

	cluster.Nodes[cluster.LocalNode.Properties.NodeIp] = cluster.LocalNode

	cluster.connectToSeeds()

	time.Sleep(15 * time.Second)

	go cluster.updateResources()

	go cluster.echo()

	time.Sleep(5 * time.Second)

	go cluster.checkNodes()

	return nil
}

func (cluster *Cluster) connectToSeeds() {
	cluster.withLock("connect to seed", func() error {
		for _, seed := range helpers.GetCfg().SeedNodes {
			node := cluster.NewEmptyNode(seed)
			cluster.LocalNode.joinTo(node)
		}

		return nil
	})
}

func (cluster *Cluster) ListNodes() []string {
	var nodes = make([]string, 0, 3)

	for _, node := range cluster.Nodes {
		nodes = append(nodes, node.Properties.NodeIp)
	}

	return nodes
}

func (cluster *Cluster) GetNode(name string) (*Node, *logger.Error) {
	node, ok := cluster.Nodes[name]
	if !ok {
		return nil, helpers.Logger.ErrorF("node not found")
	}

	return node, nil
}

func (cluster *Cluster) GetRandNode() (*Node, *logger.Error) {
	for _, node := range cluster.Nodes {
		if node.Properties.NodeIp != cluster.LocalNode.Properties.NodeIp {
			return node, nil
		}
	}

	return nil, helpers.Logger.ErrorF("there is not any reachable node")
}

func (cluster *Cluster) NewEmptyNode(ip string) *Node {
	var newNode *Node

	for _, nodeIP := range cluster.ListNodes() {
		if nodeIP == ip {
			newNode, _ = cluster.GetNode(ip)
		}
	}

	if newNode == nil {
		newNode = &Node{
			Properties: &NodeProperties{
				NodeIp: ip,
				Status: "new",
			},
			tasks: make(chan *Task, 100),
			mutex: make(chan struct{}, 1),
		}

		cluster.Nodes[newNode.Properties.NodeIp] = newNode
	}

	return newNode
}

func (cluster *Cluster) NewNode(properties *NodeProperties) *Node {
	var newNode *Node

	for _, nodeIP := range cluster.ListNodes() {
		if nodeIP == properties.NodeIp {
			newNode, _ = cluster.GetNode(properties.NodeIp)
		}
	}

	if newNode == nil {
		newNode = &Node{
			Properties: properties,
			tasks:      make(chan *Task, 100),
			mutex:      make(chan struct{}, 1),
		}

		cluster.Nodes[newNode.Properties.NodeIp] = newNode

		cluster.LocalNode.joinTo(newNode)
	}

	return newNode
}

func (cluster *Cluster) updateResources() {
	for {
		cluster.LocalNode.withLock("sending local resources update", func() error {
			cpu := runtime.NumCPU()
			cluster.LocalNode.Properties.Cores = int32(cpu)

			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)
			cluster.LocalNode.Properties.Memory = int64(mem.HeapSys)
			cluster.LocalNode.Properties.MemoryInUse = int64(mem.HeapInuse)

			cluster.LocalNode.Properties.Timestamp = time.Now().UTC().UnixMilli()

			cluster.withLock("sending resources update", func() error {
				for _, node := range cluster.Nodes {
					if node.Properties.NodeIp == cluster.LocalNode.Properties.NodeIp {
						continue
					}
					node.withLock("sending resources update", func() error {
						cluster.LocalNode.updateTo(node)

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

func (cluster *Cluster) echo() {
	for {
		cluster.withLock("sending echo", func() error {
			for _, node := range cluster.Nodes {
				ping := &Ping{
					Timestamp: time.Now().UTC().UnixMilli(),
					NodeIp:    cluster.LocalNode.Properties.NodeIp,
				}
				node.withLock("sending echo", func() error {
					pong, err := node.client().Echo(context.Background(), ping)
					if err != nil {
						e := helpers.Logger.ErrorF("cannot send ping to %s: %v", node.Properties.NodeIp, err)
						if e.Is("node not found") {
							cluster.LocalNode.joinTo(node)
							return nil
						} else {
							node.setUnhealthy(err.Error())
							return nil
						}
					}

					if pong.PongTimestamp-pong.PingTimestamp > 30000 {
						helpers.Logger.ErrorF("latency to %s is too high", node.Properties.NodeIp)
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
