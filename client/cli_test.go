package client

import (
	"context"
	"log"

	"github.com/eqweqr/chater/server"
)

func serv() {
	ctx := context.Background()
	serv := server.NewServer("0.0.0.0:6379")
	err := serv.Run(ctx)
	if err != nil {
		log.Print("cannot load server")
	}
}
