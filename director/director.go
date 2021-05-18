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
			Port: int32(openMatchMMFHostPort),
			Type: pb.FunctionConfig_GRPC,
		},
		Profile: &matchProfile,
	}

	// Fetch matches
	stream, err := be.FetchMatches(ctx, &matchesRequest)
	if err != nil {
		log.Fatalf("failed to get available stream client %v", err)
	}

	resp, err := stream.Recv()
	if err != nil {
		if err != io.EOF {
			log.Fatalf("Fetch match for profile failed %v", err)
		}

		log.Print("EOF while getting response from stream (probably no tickets for requested match profile)")
		return
	}

	match := resp.GetMatch()
	// -------------------------------------

	// Assign matches
	tickets := match.GetTickets()

	log.Printf("tickets present in match %v", tickets)
	log.Printf("tickets len %v", len(tickets))

	var ticketIDs []string
	for i := range tickets {
		ticketIDs[i] = tickets[i].Id
	}

	assignTickets := pb.AssignTicketsRequest{
		Assignments: []*pb.AssignmentGroup{
			{
				TicketIds: ticketIDs,
				Assignment: &pb.Assignment{
					// Fake connect ip address
					// We can make use of Assignment.Extensions to send customised information
					Connection: "192.168.0.111:2222",
				},
			},
		},
	}

	if _, err = be.AssignTickets(ctx, &assignTickets); err != nil {
		log.Fatalf("AssignTickets failed for match %v, got %v", match.GetMatchId(), err)
	}

	log.Printf("Assigned match to %v", match.GetMatchId())
	// -------------------------------------
}
