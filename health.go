package clustering

import (
	"time"

	"github.com/threatwinds/clustering/helpers"
)

func (cluster *Cluster) checkNodes() {
	time.Sleep(60 * time.Second)
	for {
		cluster.withLock("checking node health", func() error {
			for name, node := range cluster.Nodes {
				removeDelay := time.Now().UTC().Add(-120 * time.Second).UnixMilli()
				unhealthyDelay := time.Now().UTC().Add(-30 * time.Second).UnixMilli()
				
				node.withLock("checking node health", func() error {
					if node.Properties == nil {
						return nil
					}
					
					if node.Properties.Status != "new" && node.LastPing < removeDelay {
						delete(cluster.Nodes, name)
					}
	
					if node.Properties.Status == "healthy" && node.LastPing < unhealthyDelay {
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

func (node *Node) setUnhealthy(cause string) {
	helpers.Logger.ErrorF("node %s is unhealthy becasue: %s", node.Properties.NodeIp, cause)

	node.Latency = -1
	node.Properties.Status = "unhealthy"
}

func (node *Node) setHealthy(now, senderTime int64) {
	if node.Properties.Status != "healthy" {
		helpers.Logger.LogF(200, "node %s is now healthy", node.Properties.NodeIp)
		go node.startSending()
	}

	node.Latency = now - senderTime
	node.LastPing = senderTime
	node.Properties.Status = "healthy"
}
