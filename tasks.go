package clustering

import (
	"github.com/threatwinds/clustering/helpers"
	"github.com/threatwinds/logger"
)

// BroadcastTask broadcasts a task to all nodes in the cluster.
func (cluster *cluster) BroadcastTask(task *Task) {
	_ = cluster.EnqueueTask(task, len(cluster.nodes))
}

// EnqueueTask enqueues a task to be performed by a specified number of nodes in the cluster.
// It returns an error if there are not enough nodes available to perform the task.
func (cluster *cluster) EnqueueTask(task *Task, inNodes int) *logger.Error {
	if len(cluster.nodes) < inNodes {
		return helpers.Logger.ErrorF("not enough nodes to perform task")
	}

	alreadyAssigned := make(map[string]bool)

	for i := 0; i < inNodes; i++ {
		for _, node := range cluster.nodes {
			if node.properties.Status == "unhealthy" {
				continue
			}

			if _, ok := alreadyAssigned[node.properties.NodeIp]; !ok {
				if node.properties.Cores*100 < node.properties.RunningThreads {
					helpers.Logger.ErrorF("node %s is CPU overloaded", node.properties.NodeIp)
					continue
				}

				if node.properties.Memory-node.properties.MemoryInUse < 50 {
					helpers.Logger.ErrorF("node %s is memory overloaded", node.properties.NodeIp)
					continue
				}

				node.tasks <- task

				helpers.Logger.LogF(100, "assigned task to %s", node.properties.NodeIp)

				alreadyAssigned[node.properties.NodeIp] = true
			}
		}
	}

	if len(alreadyAssigned) < inNodes {
		return helpers.Logger.ErrorF("not enough nodes available to perform task")
	}

	return nil
}
