package clustering

import (
	"sync"
	"time"
	"runtime"
	"context"

	"github.com/threatwinds/clustering/helpers"
	"github.com/threatwinds/logger"
)

type Cluster struct {
	LocalNode    *Node
	Nodes        map[string]*Node
	callBackDict map[string]func(task *Task) (*Result, error)
	mutex 	     sync.RWMutex
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

func (cluster *Cluster) Start(callBackDict map[string]func(task *Task) (*Result, error)) *logger.Error {
	helpers.Logger.LogF(200, "starting cluster")

	var e *logger.Error

	cluster.callBackDict = callBackDict

	cluster.LocalNode = new(Node)

	cluster.LocalNode.Latency = -1

	cluster.LocalNode.tasks = make(chan *Task, 100)

	cluster.LocalNode.Properties = new(NodeProperties)

	cluster.LocalNode.Properties.NodeIp, e = helpers.GetMainIP()
	if e != nil {
		return e
	}

	cluster.LocalNode.Properties.UpSince = time.Now().UTC().UnixMilli()	

	cluster.LocalNode.Properties.Status = "new"

	cluster.LocalNode.Properties.DataCenter = helpers.GetCfg().DataCenter
	
	cluster.mutex.Lock()
	cluster.Nodes[cluster.LocalNode.Properties.NodeIp] = cluster.LocalNode
	cluster.mutex.Unlock()

	cluster.connectToSeeds()

	time.Sleep(15 * time.Second)

	go cluster.updateResources()
	
	go cluster.echo()

	time.Sleep(5 * time.Second)

	go cluster.checkNodes()

	return nil
}

func (cluster *Cluster) connectToSeeds() {
	for _, seed := range helpers.GetCfg().SeedNodes{
		node := cluster.NewEmptyNode(seed)
		cluster.LocalNode.joinTo(node)
	}
}

func (cluster *Cluster) ListNodes() []string {
	cluster.mutex.RLock()
	defer cluster.mutex.RUnlock()
	
	var nodes = make([]string, 0, 3)

	for _, node := range cluster.Nodes {
		nodes = append(nodes, node.Properties.NodeIp)
	}

	return nodes
}

func (cluster *Cluster) GetNode(name string) (*Node, *logger.Error) {
	cluster.mutex.RLock()
	defer cluster.mutex.RUnlock()

	node, ok := cluster.Nodes[name]
	if !ok{
		return nil, helpers.Logger.ErrorF("node not found")
	}

	return node, nil
}

func (cluster *Cluster) GetRandNode() (*Node, *logger.Error) {
	cluster.mutex.RLock()
	defer cluster.mutex.RUnlock()

	for _, node := range cluster.Nodes {
		if node.Properties.NodeIp != cluster.LocalNode.Properties.NodeIp{
			return node, nil
		}
	}

	return nil, helpers.Logger.ErrorF("there is not any reachable node")
}

func (cluster *Cluster) NewEmptyNode(ip string) *Node {
	var newNode *Node

	for _, nodeIP := range cluster.ListNodes(){
		if nodeIP == ip{
			newNode, _ = cluster.GetNode(ip)
		}
	} 
	
	if newNode == nil {
		cluster.mutex.Lock()
		defer cluster.mutex.Unlock()
		
		newNode = &Node{
			Properties: &NodeProperties{
				NodeIp: ip,
				Status: "new",
			},
			tasks: make(chan *Task, 100),
		}

		cluster.Nodes[newNode.Properties.NodeIp] = newNode
	}

	return newNode
}

func (cluster *Cluster) NewNode(properties *NodeProperties) *Node {
	var newNode *Node

	for _, nodeIP := range cluster.ListNodes(){
		if nodeIP == properties.NodeIp{
			newNode, _ = cluster.GetNode(properties.NodeIp)
		}
	} 

	if newNode == nil {
		cluster.mutex.Lock()
		defer cluster.mutex.Unlock()

		newNode = &Node{
			Properties: properties,
			tasks: make(chan *Task, 100),
		}

		cluster.Nodes[newNode.Properties.NodeIp] = newNode
		
		cluster.LocalNode.joinTo(newNode)
	}

	return newNode
}


func (cluster *Cluster) updateResources(){
	for{
		cluster.LocalNode.mutex.Lock()
		
		cpu := runtime.NumCPU()
		cluster.LocalNode.Properties.Cores = int32(cpu)
		
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		cluster.LocalNode.Properties.Memory = int64(mem.HeapSys)
		cluster.LocalNode.Properties.MemoryInUse = int64(mem.HeapInuse)
		
		cluster.LocalNode.Properties.Timestamp = time.Now().UTC().UnixMilli()
		
		cluster.LocalNode.mutex.Unlock()

		for _, node := range cluster.Nodes {
			if node.Properties.Status != "healthy"{
				continue
			}

			_, err := node.client().UpdateNode(context.Background(), cluster.LocalNode.Properties)
			if err != nil {
				e := helpers.Logger.ErrorF("cannot send resources update to %s: %v", node.Properties.NodeIp, err)
				if e.Is("node not found"){
					cluster.LocalNode.joinTo(node)
					continue
				}else{
					node.setUnhealthy(err.Error())
					continue
				} 
			}
		}

		time.Sleep(10 * time.Second)
	}
}

func (cluster *Cluster) echo() {
	for {
		for _, node := range cluster.Nodes{
			ping := &Ping{
				Timestamp: time.Now().UTC().UnixMilli(),
				NodeIp:    cluster.LocalNode.Properties.NodeIp,
			}
		
			pong, err := node.client().Echo(context.Background(), ping)
			if err != nil {
				e := helpers.Logger.ErrorF("cannot send ping to %s: %v", node.Properties.NodeIp, err)
				if e.Is("node not found"){
					cluster.LocalNode.joinTo(node)
					continue
				}else{
					node.setUnhealthy(err.Error())
					continue
				} 
			}
		
			if pong.PongTimestamp-pong.PingTimestamp > 30000 {
				helpers.Logger.ErrorF("latency to %s is too high", node.Properties.NodeIp)
				node.setUnhealthy("latency too high")
			}
		}

		time.Sleep(1 * time.Second)
	}
}
