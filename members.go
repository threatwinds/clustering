package clustering

import (
	context "context"
	"fmt"
	"github.com/threatwinds/clustering/helpers"
	"log"
	"strings"
	"sync"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type member struct {
	Name          string
	IP            string
	Latency       float64
	Local         bool
	Master        bool
	Status        string
	LastKeepAlive int64
	UpTimestamp   int64
}

var members = make(map[string]*member)
var membersMutex sync.RWMutex
var pool sync.WaitGroup

func discover() {
	time.Sleep(10 * time.Second)
	for {
		log.Println("discovering cluster nodes")

		segments := strings.Split(localNode.Ip, ".")

		network := fmt.Sprintf("%s.%s.%s", segments[0], segments[1], segments[2])

		for i := 1; i <= 254; i++ {
			host := fmt.Sprintf("%s.%d", network, i)
			pool.Add(1)
			go func() {
				err := register(localNode, host)
				if err != nil {
					if !strings.Contains(err.Error(), "context deadline exceeded") &&
						!strings.Contains(err.Error(), "host is down") &&
						!strings.Contains(err.Error(), "connection refused") &&
						!strings.Contains(err.Error(), "no route to host") &&
						!strings.Contains(err.Error(), "error reading server preface: EOF") {
						log.Printf("error sending registration request to %s: %s", host, err.Error())
					}
				}
				pool.Done()
			}()
		}
		pool.Wait()
		time.Sleep(10 * time.Second)
	}

}

func setMaster(name string) {
	if name == "" {
		return
	}

	if name == localNode.Name {
		localNode.Master = true
	} else {
		localNode.Master = false
	}

	membersMutex.Lock()

	for _, m := range members {
		m.Master = false
		if m.Name == name {
			m.Master = true
		}
	}

	membersMutex.Unlock()
}

func connect(ip string) (*connection, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", ip, helpers.GetCfg().ClusterPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return &connection{}, err
	}

	cnn := &connection{cnn: conn}

	connectionsMutex.Lock()

	connections[ip] = cnn

	connectionsMutex.Unlock()

	go keepAlive(cnn)

	return cnn, nil
}

func getMember(name string) (*member, error) {
	membersMutex.RLock()
	m, ok := members[name]
	membersMutex.RUnlock()
	if !ok {
		return nil, fmt.Errorf("member %s does not exists", name)
	}

	return m, nil
}

func keepAlive(c *connection) {
	ctx := context.Background()

	r, err := c.Client().Ping(ctx)
	if err != nil {
		return
	}

	var se int

	for {
		pong := Pong{Timestamp: time.Now().UTC().UnixMicro(), Name: localNode.Name}
		if err := r.Send(&pong); err != nil {
			log.Printf("error sending ping to %s: %s", c.cnn.Target(), err.Error())
			se++
			if se >= 2 {
				return
			}
		} else {
			se = 0
		}
		time.Sleep(10 * time.Second)
	}
}

func getConnection(ip string) (*connection, error) {
	connectionsMutex.RLock()
	connection, ok := connections[ip]
	connectionsMutex.RUnlock()

	if !ok {
		return connect(ip)
	}

	return connection, nil
}

func register(m *Member, ip string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, err := getConnection(ip)
	if err != nil {
		return err
	}

	_, err = conn.Client().Register(ctx, m)
	if err != nil {
		disconnect(ip)
		return err
	}

	return nil
}

func disconnect(ip string) {
	c, err := getConnection(ip)
	if err != nil {
		return
	}

	connectionsMutex.Lock()
	if c == nil {
		return
	}

	if c.cnn == nil {
		return
	}

	_ = c.cnn.Close()
	delete(connections, ip)
	connectionsMutex.Unlock()
}
