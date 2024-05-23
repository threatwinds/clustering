package clustering

import (
	"fmt"
	"time"

	"github.com/threatwinds/clustering/helpers"
	"github.com/threatwinds/logger"
)

// BroadcastTask broadcasts a task to all nodes in the cluster.
func (cluster *cluster) BroadcastTask(task *Task) int {
	nodes, _ := cluster.EnqueueTask(task, len(cluster.nodes))

	return nodes
}

// EnqueueTask enqueues a task to be performed by a specified number of nodes in the cluster.
// It returns an error if there are not enough nodes available to perform the task.
func (cluster *cluster) EnqueueTask(task *Task, inNodes int) (int, *logger.Error) {
	var assigned int

	e := cluster.withLock("EnqueueTask", func() error {
		if len(cluster.healthyNodes()) < inNodes {
			return fmt.Errorf("not enough nodes to perform task")
		}

		alreadyAssigned := make(map[string]bool)

		for i := 0; i < inNodes; i++ {
			for _, node := range cluster.healthyNodes() {
				if _, ok := alreadyAssigned[node.properties.NodeIp]; !ok {
					if node.properties.Cores*int32(len(cluster.nodes)*20) < node.properties.RunningThreads {
						helpers.Logger.ErrorF("node %s is CPU overloaded", node.properties.NodeIp)
						continue
					}

					if node.properties.Memory-node.properties.MemoryInUse < 50 {
						helpers.Logger.ErrorF("node %s is memory overloaded", node.properties.NodeIp)
						continue
					}

					select {
					case node.tasks <- task:
						alreadyAssigned[node.properties.NodeIp] = true
					case <-time.After(5 * time.Second):
						node.setUnhealthy("time out assigning task")
					}
				}
			}
		}

		if len(alreadyAssigned) < inNodes {
			assigned = len(alreadyAssigned)
			return fmt.Errorf("not enough nodes available to perform task")
		}

		assigned = len(alreadyAssigned)

		return nil
	})

	return assigned, e
}
