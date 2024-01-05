package client

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"

	pb "github.com/eqweqr/chater/protoc"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type client struct {
	pb.ChatClient
	Name string
}

func NewClient() (*client, error) {
	id, err := uuid.NewRandom()
	if err != nil {

	}
	return &client{
		Name: id.String(),
	}, nil
}

func (c *client) Run(ctx context.Context) error {
	connCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	conn, err := grpc.DialContext(connCtx, "0.0.0.0:8081", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return errors.WithMessage(err, "unable to connect")
	}
	defer conn.Close()

	c.ChatClient = pb.NewChatClient(conn)

	log.Print("user connected to chat")
	err = c.stream(ctx)

	return errors.WithMessage(err, "stream error")

}

func (c *client) stream(ctx context.Context) error {
	// md := metadata.New(map[string]string{"redis": "redis"})
	// ctx = metadata.NewOutgoingContext(ctx, md)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	client, err := c.ChatClient.Stream(ctx)
	if err != nil {
		return err
	}
	defer client.CloseSend()

	log.Print("client connected to stream")

	go c.send(client)
	return c.receive(client)
}

func (c *client) receive(sc pb.Chat_StreamClient) error {
	for {
		resp, err := sc.Recv()

		if s, ok := status.FromError(err); ok && s.Code() == codes.Canceled {
			log.Print("stream canceled internal causes")
			return nil
		} else if err == io.EOF {
			log.Print("stream closed by server")
			return nil
		} else if err != nil {
			return err
		}
		log.Printf("%s: %s", resp.Auth, resp.Message)
	}
}

func (c *client) send(client pb.Chat_StreamClient) {
	sc := bufio.NewScanner(os.Stdin)
	sc.Split(bufio.ScanLines)

	for {
		select {
		case <-client.Context().Done():
			log.Print("client close connection")
		default:
			if sc.Scan() {
				if err := client.Send(&pb.StreamRequest{
					Message: sc.Text(),
				}); err != nil {
					log.Print("failed to send message")
					return
				}
			} else {
				log.Print("scann input failed")
				return
			}
		}
	}
}
