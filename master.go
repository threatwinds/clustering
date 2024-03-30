package clustering

import (
	"context"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"
)

func sendMasterVote() {
	time.Sleep(10 * time.Second)
	for {
		membersMutex.RLock()
		tmp := members
		membersMutex.RUnlock()

		var vote = new(Vote)
		var latency = float64(10000)
		var old string
		var oldTimestamp = time.Now().UTC().UnixMicro()

		for _, node := range tmp {
			if node.Status != "healthy" {
				continue
			}

			if latency > node.Latency {
				vote.Name = node.Name
				latency = node.Latency
			}

			if oldTimestamp > node.UpTimestamp {
				oldTimestamp = node.UpTimestamp
				old = node.Name
			}
		}

		ctx := context.Background()

		votesServer, err := getMember(old)
		if err != nil {
			continue
		}

		conn, err := getConnection(votesServer.IP)
		if err != nil {
			continue
		}

		_, err = conn.Client().MasterVote(ctx, vote)
		if err != nil {
			continue
		}

		time.Sleep(10 * time.Second)
	}
}

var votes = make(map[string]int)
var votesMutex sync.RWMutex

func (s *Server) MasterVote(ctx context.Context, in *Vote) (*emptypb.Empty, error) {
	votesMutex.Lock()

	_, ok := votes[in.Name]
	if !ok {
		votes[in.Name] = 1
	} else {
		votes[in.Name] += 1
	}

	votesMutex.Unlock()

	return new(emptypb.Empty), nil
}

func (s *Server) ElectedMaster(ctx context.Context, in *Vote) (*emptypb.Empty, error) {
	setMaster(in.Name)
	return new(emptypb.Empty), nil
}

func sendElectedMaster(vote *Vote) {
	membersMutex.RLock()
	tmp := members
	membersMutex.RUnlock()

	for _, node := range tmp {
		if node.Status != "healthy" {
			continue
		}

		ctx := context.Background()

		member, err := getMember(node.Name)
		if err != nil {
			continue
		}

		conn, err := getConnection(member.IP)
		if err != nil {
			continue
		}

		_, err = conn.Client().ElectedMaster(ctx, vote)
		if err != nil {
			continue
		}
	}
}

func processVotes() {
	time.Sleep(30 * time.Second)
	var current string
	var amount int
	for {
		votesMutex.RLock()
		for name, vote := range votes {
			if amount < vote {
				current = name
				amount = vote
			}
		}

		sendElectedMaster(&Vote{Name: current})

		votes = make(map[string]int)
		votesMutex.RUnlock()

		time.Sleep(5 * time.Minute)
	}
}
