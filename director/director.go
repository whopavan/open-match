package main

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"open-match.dev/open-match/pkg/pb"
)

var (
	openMatchBackendEndpoint = "something"
)

func main() {
	conn, err := grpc.Dial("open-match-backend.open-match.svc.cluster.local:50505", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	var (
		ctx = context.Background()
		be  = pb.NewBackendServiceClient(conn)
	)

	matches := pb.FetchMatchesRequest{}
	log.Printf("%v %v %v", ctx, be, matches)
}
