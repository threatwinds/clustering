package clustering

import (
	context "context"
	"log"
)

func (s *Server) Receiver(srv Cluster_ReceiverServer, callback func(msg *Message) error) error {
	msg, err := srv.Recv()
	if err != nil {
		return err
	}

	return callback(msg)
}

func SendThread(c *connection, queue chan *Message) {
	ctx := context.Background()

	client, err := c.Client().Receiver(ctx)
	if err != nil {
		return
	}

	for {
		msg := <-queue
		if err := client.Send(msg); err != nil {
			log.Printf("error sending msg to %s: %s", c.cnn.Target(), err.Error())
			return
		}
	}
}
