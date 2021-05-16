package main

import (
	"context"
	"io"
	"log"

	"google.golang.org/grpc"
	"open-match.dev/open-match/pkg/pb"
)

var (
	openMatchBackendEndpoint = "192.168.49.2:31855"

	openMatchMMFHostName = "192.168.0.117"
	openMatchMMFPort     = int32(50506)
)

func main() {
	conn, err := grpc.Dial(openMatchBackendEndpoint, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	var (
		ctx = context.Background()
		be  = pb.NewBackendServiceClient(conn)
	)

	matchProfile := pb.MatchProfile{
		Name: "deathmatch",
		Pools: []*pb.Pool{
			{
				TagPresentFilters: []*pb.TagPresentFilter{
					{
						Tag: "deathmatch",
					},
				},
			},
		},
	}

	matchesRequest := pb.FetchMatchesRequest{
		Config: &pb.FunctionConfig{
			Host: openMatchMMFHostName,
			Port: openMatchMMFPort,
			Type: pb.FunctionConfig_GRPC,
		},
		Profile: &matchProfile,
	}

	// fetch matches
	stream, err := be.FetchMatches(ctx, &matchesRequest)
	if err != nil {
		log.Fatalf("failed to get available stream client %v", err)
	}

	resp, err := stream.Recv()
	if err == io.EOF {
		log.Printf("EOF in resp")
		return
	}

	if err != nil {
		log.Fatalf("Pull match failed %v", err)
	}

	matchToAssign := resp.GetMatch()
	// -------------------------------------

	// assign matches
	tickets := matchToAssign.GetTickets()

	var ticketIDs []string
	for i := range tickets {
		ticketIDs[i] = tickets[i].Id
	}

	assignTicket := pb.AssignTicketsRequest{
		Assignments: []*pb.AssignmentGroup{
			{
				TicketIds: ticketIDs,
				Assignment: &pb.Assignment{
					// dummy connect ip address
					Connection: "192.168.0.111:2222",
				},
			},
		},
	}

	if _, err = be.AssignTickets(ctx, &assignTicket); err != nil {
		log.Fatalf("AssignTickets failed for match %v, got %v", matchToAssign.GetMatchId(), err)
	}

	log.Printf("Assigned server %v to match %v", conn, matchToAssign.GetMatchId())
	// -------------------------------------
}
