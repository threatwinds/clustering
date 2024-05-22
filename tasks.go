package clustering

import (
	"github.com/threatwinds/clustering/helpers"
	"github.com/threatwinds/logger"
)

func (cluster *cluster) BroadcastTask(task *Task) {
	nodes := cluster.listNodes()
	_ = cluster.EnqueueTask(task, len(nodes))
}

func (cluster *cluster) EnqueueTask(task *Task, inNodes int) *logger.Error {
	nodes := cluster.listNodes()

	if len(nodes) < inNodes {
		return helpers.Logger.ErrorF("not enough nodes to perform task")
	}

	alreadyAssigned := make(map[string]bool)

	for i := 0; i < inNodes; i++ {
		for _, nodeIP := range nodes {
			node, e := cluster.getNode(nodeIP)
			if e != nil {
				return helpers.Logger.ErrorF("error getting node: %v", e)
			}

			if _, ok := alreadyAssigned[node.properties.NodeIp]; !ok {
				if node.properties.Cores*100 < node.properties.RunningThreads {
					continue
				}

				if node.properties.Memory-node.properties.MemoryInUse < 50 {
					continue
				}

				node.tasks <- task
				alreadyAssigned[node.properties.NodeIp] = true
			}
		}
	}

	if len(alreadyAssigned) < inNodes {
		return helpers.Logger.ErrorF("not enough nodes available to perform task")
	}

	return nil
}
