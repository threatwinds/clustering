package clustering

import (
	"time"

	"github.com/threatwinds/clustering/helpers"
)

// checkNodes periodically checks the health of all nodes in the cluster.
// It removes nodes that haven't sent a ping in the last 120 seconds,
// and marks nodes with high latency as unhealthy.
func (cluster *cluster) checkNodes() {
	time.Sleep(60 * time.Second)
	for {
		var deleteQueue = make([]string, 0, 1)

		cluster.withLock("check cluster health", func() error {
			for name, node := range cluster.nodes {
				unhealthyDelay := time.Now().UTC().Add(-10 * time.Second).UnixMilli()
				removeDelay := time.Now().UTC().Add(-30 * time.Second).UnixMilli()

				if node.properties == nil {
					continue
				}

				if node.properties.Status == "healthy" && node.lastPing < unhealthyDelay {
					node.setUnhealthy("high latency")
				}

				if node.properties.Status != "new" && node.lastPing < removeDelay {
					deleteQueue = append(deleteQueue, name)
				}

				helpers.Logger.LogF(200, "node %s status is %s", node.properties.NodeIp, node.properties.Status)
			}

			for _, name := range deleteQueue {
				delete(cluster.nodes, name)
			}

			return nil
		})

		time.Sleep(60 * time.Second)
	}
}

// setUnhealthy marks the node as unhealthy with the given cause.
func (node *node) setUnhealthy(cause string) {
	if node.properties.Status == "unhealthy" {
		return
	}

	helpers.Logger.ErrorF("node %s is unhealthy: %s", node.properties.NodeIp, cause)
	node.latency = -1
	node.properties.Status = "unhealthy"
}

// setHealthy updates the node's status to healthy and sets the latency and last ping time.
func (node *node) setHealthy() {
	if node.properties.Status == "healthy" {
		return
	}

	helpers.Logger.LogF(200, "node %s is now healthy", node.properties.NodeIp)
	node.properties.Status = "healthy"
}
