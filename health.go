package clustering

import (
	"log"
	"time"
)

func checkNodes() {
	time.Sleep(10 * time.Second)
	for {
		membersMutex.RLock()
		tmp := members
		membersMutex.RUnlock()

		for _, node := range tmp {
			log.Println(*node)
			delay := time.Now().UTC().Add(-25 * time.Second).UnixMicro()
			if node.LastKeepAlive < delay && node.Status == "healthy" {
				setUnhealthy(node.Name)
			}
		}

		time.Sleep(10 * time.Second)
	}
}

func setUnhealthy(name string) {
	m, err := getMember(name)
	if err != nil {
		return
	}

	membersMutex.Lock()

	m.Latency = -1
	m.Status = "unhealthy"
	m.Master = false

	members[name] = m

	disconnect(m.IP)

	membersMutex.Unlock()
	if m.Local {
		log.Fatalf("self connection lost")
	}
}

func setHealthy(name string, now, senderTime int64) {
	m, err := getMember(name)
	if err != nil {
		return
	}

	membersMutex.Lock()

	m.Latency = float64(now-senderTime) / 1000
	m.LastKeepAlive = senderTime
	m.Status = "healthy"

	members[name] = m

	membersMutex.Unlock()

	log.Printf("received keep alive from %s, %v milliseconds", name, m.Latency)
}
