package main

import (
	"context"
	"io"
	"log"
	"os"
	"strconv"

	"google.golang.org/grpc"
	"open-match.dev/open-match/pkg/pb"
)

func main() {
	openMatchBackendEndpoint, envSet := os.LookupEnv("OPEN_MATCH_BACKEND_ENDPOINT")
	if !envSet {
		log.Fatalf("Open match env OPEN_MATCH_BACKEND_ENDPOINT not set")
	}

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

	openMatchMMFHostName, envSet := os.LookupEnv("OPEN_MATCH_MATCHFUNCTION_HOSTNAME")
	if !envSet {
		log.Fatalf("Open match env OPEN_MATCH_MATCHFUNCTION_HOSTNAME not set")
	}

	openMatchMMFHostPortStr, envSet := os.LookupEnv("OPEN_MATCH_MATCHFUNCTION_HOSTPORT")
	if !envSet {
		log.Fatalf("Open match env OPEN_MATCH_MATCHFUNCTION_HOSTPORT not set")
	}
	openMatchMMFHostPort, err := strconv.ParseInt(openMatchMMFHostPortStr, 10, 32)
	if err != nil {
		log.Fatalf("Parsing OPEN_MATCH_MATCHFUNCTION_HOSTPORT failed %v", err)
	}

	log.Printf("MMF host %s port %d", openMatchMMFHostName, openMatchMMFHostPort)

	matchesRequest := pb.FetchMatchesRequest{
		Config: &pb.FunctionConfig{
			Host: openMatchMMFHostName,
			Port: int32(openMatchMMFHostPort),
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
		log.Fatalf("EOF while getting response from stream")
		return
	}
	if err != nil {
		log.Fatalf("Fetch match for profile failed %v", err)
	}

	matchToAssign := resp.GetMatch()
	// -------------------------------------

	// assign matches
	tickets := matchToAssign.GetTickets()

	log.Printf("tickets %v", tickets)
	log.Printf("tickets %v", len(tickets))

	// var ticketIDs []string
	// for i := range tickets {
	// 	ticketIDs[i] = tickets[i].Id
	// }

	assignTicket := pb.AssignTicketsRequest{
		Assignments: []*pb.AssignmentGroup{
			{
				TicketIds: []string{tickets[0].Id},
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
