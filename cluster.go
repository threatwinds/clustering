package clustering

import (
	context "context"
	"github.com/threatwinds/clustering/helpers"
	"log"
	"os"
	sync "sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var localNode = new(Member)

type connection struct {
	cnn *grpc.ClientConn
}

func (c *connection) Client() ClusterClient {
	return NewClusterClient(c.cnn)
}

var connections = make(map[string]*connection)
var connectionsMutex sync.RWMutex

func init() {
	var err error
	log.Println("cluster starting")

	localNode.Name, err = os.Hostname()
	if err != nil {
		log.Fatalln("cannot get hostname")
	}

	localNode.Ip, err = helpers.GetMainIP()
	if err != nil {
		log.Fatalln("cannot get main IP")
	}

	localNode.UpTimestamp = time.Now().UTC().UnixMicro()

	go discover()

	go checkNodes()

	go sendMasterVote()

	go processVotes()
}

type Server struct {
	UnimplementedClusterServer
}

func (s *Server) Register(ctx context.Context, in *Member) (*emptypb.Empty, error) {
	_, err := getMember(in.Name)
	if err != nil {
		log.Printf("registration request from: %v", in.GetName())

		membersMutex.Lock()
		members[in.Name] = &member{
			Name:    in.Name,
			IP:      in.Ip,
			Latency: -1,
			Local: func() bool {
				return localNode.Name == in.Name
			}(),
			Master: in.Master,
			Status: "new",
		}
		membersMutex.Unlock()

		log.Printf("discovered node %s at %s", in.Name, in.Ip)
	}

	return new(emptypb.Empty), nil
}

func (s *Server) Ping(srv Cluster_PingServer) error {
	var re int
	var name string

	for {
		req, err := srv.Recv()
		if err != nil {
			re++
			if re >= 2 {
				setUnhealthy(name)
				return err
			}
			time.Sleep(1 * time.Second)
		} else {
			re = 0
			name = req.Name
			setHealthy(name, time.Now().UTC().UnixMicro(), req.Timestamp)
		}
	}
}
