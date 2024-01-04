package main

import (
	"context"
	"log"

	"github.com/eqweqr/chater/server"
)

// func closeSignal(rdb redis.Client) error {
// 	signChan := make(chan os.Signal, 1)
// 	signal.Notify(signChan, syscall.SIGTERM, syscall.SIGINT)
// 	log.Printf("Waiting for close signal")
// 	<-signChan
// 	log.Printf("Signal received!")
// 	signal.Stop(signChan)
// 	close(signChan)
// 	if err := rdb.Close(); err != nil {
// 		log.Printf("Error with closing redis db: %w", err)
// 		return fmt.Errorf("cannot close rdb(%s): %w", rdb, err)
// 	}
// 	log.Printf("%s was closed successfuly", rdb)
// 	return nil
// }

// type MessageServer struct {
// 	RedisClient *redis.Client
// }

func main() {
	srv := server.NewServer("0.0.0.0:6379")
	log.Fatal(srv.Run(context.Background()))
}
