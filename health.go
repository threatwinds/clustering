package clustering

import (
	"time"

	"github.com/threatwinds/clustering/helpers"
)

func (cluster *cluster) checkNodes() {
	time.Sleep(60 * time.Second)
	for {
		cluster.withLock("checking node health", func() error {
			for name, node := range cluster.nodes {
				removeDelay := time.Now().UTC().Add(-120 * time.Second).UnixMilli()
				unhealthyDelay := time.Now().UTC().Add(-30 * time.Second).UnixMilli()
				
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

func (node *node) setUnhealthy(cause string) {
	helpers.Logger.ErrorF("node %s is unhealthy: %s", node.properties.NodeIp, cause)

	node.latency = -1
	node.properties.Status = "unhealthy"
}

func (node *node) setHealthy(now, senderTime int64) {
	if node.properties.Status != "healthy" {
		helpers.Logger.LogF(200, "node %s is now healthy", node.properties.NodeIp)
		go node.startSending()
	}

	node.latency = now - senderTime
	node.lastPing = senderTime
	node.properties.Status = "healthy"
}
