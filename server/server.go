package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	pb "github.com/eqweqr/chater/protoc"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct {
	RedisClient  MessageStore
	Broadcast    chan *pb.StreamResponse
	ClientStream map[string]chan *pb.StreamResponse
	ClientNames  map[string]string
	streamMtx    sync.RWMutex

	pb.UnimplementedChatServer
}

type GRPCInterceptor struct {
}

func NewGRPCInterceptor() *GRPCInterceptor {
	return &GRPCInterceptor{}
}

// middelware for stream
func (interceptor *GRPCInterceptor) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		log.Printf("user connected", srv)
		return handler(srv, ss)
	}
}

func NewServer(address string) *server {
	rdb, err := NewRedisMessageStore(address)
	if err != nil {
		log.Print("cannot create store: %w", err)
	}

	return &server{
		RedisClient: rdb,

		Broadcast:    make(chan *pb.StreamResponse, 1000),
		ClientNames:  make(map[string]string),
		ClientStream: make(map[string]chan *pb.StreamResponse),
	}
}

func (messageServer *server) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	interceptor := NewGRPCInterceptor()
	log.Print("Starting chat server.")
	srv := grpc.NewServer(grpc.StreamInterceptor(interceptor.StreamInterceptor()))
	pb.RegisterChatServer(srv, messageServer)
	// register reflection service on grpc server
	reflection.Register(srv)

	l, err := net.Listen("tcp", "0.0.0.0:8081")
	fmt.Print(l.Addr().String())
	if err != nil {
		log.Print("cannot start server")
		return err
	}

	go messageServer.broadcast()

	go func() {
		_ = srv.Serve(l)
		cancel()
	}()

	<-ctx.Done()

	close(messageServer.Broadcast)
	log.Print("stopping server")

	srv.GracefulStop()
	return nil
}

func (messageServer *server) Stream(srv pb.Chat_StreamServer) error {
	// username, err := messageServer.extractName(srv.Context())
	// if err != nil {
	// 	return status.Error(codes.Unauthenticated, "missing username")
	// }
	id, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("cannot create new id: %w", err)
	}
	go messageServer.sendBroadcasts(srv, id.String())

	for {
		req, err := srv.Recv()
		log.Printf("message: %s", req)
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Internal, "cannot receive message")
		}

		value, err := messageServer.RedisClient.SetMessage(srv.Context(), req.Message, time.Hour)
		log.Printf("id new message: %s", value)

		if err != nil {
			return status.Error(codes.Internal, "cannot save message to redis")
		}
		messageServer.Broadcast <- &pb.StreamResponse{
			Message: req.Message,
			Auth:    id.String(),
		}
	}

	<-srv.Context().Done()
	return srv.Context().Err()

}

// func (messageServer *server) extractName(ctx context.Context) (string, error) {
// 	md, ok := metadata.FromIncomingContext(ctx)
// 	if !ok || len(md["username"][0]) == 0 {
// 		return "", fmt.Errorf("cannot receive name from message")
// 	}
// 	return md["username"][0], nil
// }

func (messageServer *server) sendBroadcasts(srv pb.Chat_StreamServer, username string) {
	stream := messageServer.openStream(username)
	defer messageServer.closeStream(username)

	for {
		select {
		case <-srv.Context().Done():
			return
		case res := <-stream:
			if errorCode, ok := status.FromError(srv.Send(res)); ok {
				switch errorCode.Code() {
				case codes.OK:
					log.Print("message delivered to user")
				case codes.Unauthenticated, codes.Canceled, codes.DeadlineExceeded:
					log.Printf("client %s terminted connection", username)
					return
				default:
					log.Print("cannot send message: %w", errorCode.Err())
					return
				}
			}
		}
	}
}

// отправить сообщение во все каналы
func (messageServer *server) broadcast() {
	for res := range messageServer.Broadcast {
		messageServer.streamMtx.Lock()
		for _, stream := range messageServer.ClientStream {
			select {
			case stream <- res:
			default:
				log.Print("client send message at ", time.Now().Local().Format("yy.mm.dd"))
			}
		}
	}
}

// открываем поток для пользователя.
func (messageServer *server) openStream(username string) (stream chan *pb.StreamResponse) {
	stream = make(chan *pb.StreamResponse, 100)

	messageServer.streamMtx.RLock()
	messageServer.ClientStream[username] = stream
	messageServer.streamMtx.RUnlock()

	log.Printf("%s opened stream", username)
	return
}

// закрываем поток пользователя
func (messageServer *server) closeStream(username string) {
	messageServer.streamMtx.RLock()
	if stream, ok := messageServer.ClientStream[username]; ok {
		delete(messageServer.ClientStream, username)
		close(stream)
	}
	log.Printf("%s closed stream", username)
	messageServer.streamMtx.RUnlock()
}
