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
		cluster.withLock("checking node health", func() error {
			for name, node := range cluster.nodes {
				removeDelay := time.Now().UTC().Add(-30 * time.Second).UnixMilli()
				unhealthyDelay := time.Now().UTC().Add(-10 * time.Second).UnixMilli()
				
				node.withLock("checking node health", func() error {
					if node.properties == nil {
						return nil
					}
					
					if node.properties.Status != "new" && node.lastPing < removeDelay {
						delete(cluster.nodes, name)
					}
	
					if node.properties.Status == "healthy" && node.lastPing < unhealthyDelay {
						node.setUnhealthy("of high latency")
					}

					return nil
				})
			}

			return nil
		})

		time.Sleep(10 * time.Second)
	}
}

// setUnhealthy marks the node as unhealthy with the given cause.
func (node *node) setUnhealthy(cause string) {
	helpers.Logger.ErrorF("node %s is unhealthy: %s", node.properties.NodeIp, cause)

	node.latency = -1
	node.properties.Status = "unhealthy"
}

// setHealthy updates the node's status to healthy and sets the latency and last ping time.
func (node *node) setHealthy(now, senderTime int64) {
	if node.properties.Status != "healthy" {
		helpers.Logger.LogF(200, "node %s is now healthy", node.properties.NodeIp)
	}

	node.latency = now - senderTime
	node.lastPing = senderTime
	node.properties.Status = "healthy"
}
