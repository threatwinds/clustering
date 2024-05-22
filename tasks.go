package clustering

import (
	"github.com/threatwinds/clustering/helpers"
	"github.com/threatwinds/logger"
)

func (cluster *Cluster) BroadcastTask(task *Task) {
	nodes := cluster.ListNodes()
	_ = cluster.EnqueueTask(task, len(nodes))
}

func (cluster *Cluster) EnqueueTask(task *Task, inNodes int) *logger.Error {
	nodes := cluster.ListNodes()

	if len(nodes) < inNodes {
		return helpers.Logger.ErrorF("not enough nodes to perform task")
	}

	alreadyAssigned := make(map[string]bool)

	for i := 0; i < inNodes; i++ {
		for _, nodeIP := range nodes {
			node, e := cluster.GetNode(nodeIP)
			if e != nil {
				return helpers.Logger.ErrorF("error getting node: %v", e)
			}

			if _, ok := alreadyAssigned[node.Properties.NodeIp]; !ok {
				if node.Properties.Cores < node.Properties.RunningThreads {
					continue
				}

				if node.Properties.Memory-node.Properties.MemoryInUse < 50 {
					continue
				}

				node.tasks <- task
				alreadyAssigned[node.Properties.NodeIp] = true
			}
		}
	}

	if len(alreadyAssigned) < inNodes {
		return helpers.Logger.ErrorF("not enough nodes available to perform task")
	}

	return nil
}
