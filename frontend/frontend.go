package main

import (
	"context"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"open-match.dev/open-match/pkg/pb"
)

/*

	How to get end point of a service
	in yaml file set
		- Service spec.type to LoadBalancer
		- Set random clusterIP
	Apply changes in yaml file
		- kubectl apply -f 01-open-match-core.yaml
		- kubectl apply --namespace open-match -f 06-open-match-override-configmap.yaml -f 07-open-match-default-evaluator.yaml
	Check pod status, once it turns to `Running` then go to next step
		- for((i=0; ;++i)); do kubectl get pod -o wide -n open-match; sleep 5; done
	Get all services running in our namespace, here external ip should say pending
		- kubectl get services -n open-match
		example:
			open-match-backend          LoadBalancer   10.96.0.102      <pending>     50505:30970/TCP,51505:31626/TCP   6m4s
			open-match-evaluator        ClusterIP      None             <none>        50508/TCP,51508/TCP               6m2s
			open-match-frontend         LoadBalancer   10.96.0.103      <pending>     50504:30348/TCP,51504:30578/TCP   6m4s
			open-match-query            LoadBalancer   10.96.0.104      <pending>     50503:30404/TCP,51503:32568/TCP   6m4s
			open-match-redis            LoadBalancer   10.109.115.202   <pending>     6379:30444/TCP,26379:32278/TCP    6m23s
			open-match-redis-headless   LoadBalancer   10.96.0.101      <pending>     6379:31625/TCP,26379:30445/TCP    6m4s
			open-match-redis-metrics    LoadBalancer   10.108.159.34    <pending>     9121:30736/TCP                    6m23s
			open-match-swaggerui        ClusterIP      10.105.224.161   <none>        51500/TCP                         6m23s
			open-match-synchronizer     LoadBalancer   10.102.170.237   <pending>     50506:30723/TCP,51506:30405/TCP   6m23s
	Get URL for specific service
		- minikube service [service-name] --url -n open-match
		example:
			- minikube service open-match-frontend --url -n open-match
		returns:
			http://192.168.49.2:30348
			http://192.168.49.2:30578
	Copy only IP:PORT and paste it to your endpoint variable like below

*/

func main() {
	openMatchFrontendEndpoint, nv := os.LookupEnv("OPEN_MATCH_FRONTEND_ENDPOINT")
	if !nv {
		log.Fatalf("Open match env not set %v", nv)
	}

	// connect to "open match frontend"
	conn, err := grpc.Dial(openMatchFrontendEndpoint, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to Open Match, got %v", err)
	}

	defer conn.Close()
	var (
		fe  = pb.NewFrontendServiceClient(conn)
		ctx = context.Background()
	)

	// Below block should be automatically sent by a script or some sort
	newTicketRequest := pb.CreateTicketRequest{
		Ticket: &pb.Ticket{
			SearchFields: &pb.SearchFields{
				Tags: []string{"deathmatch"},
			},
			Id: "1234",
		},
	}
	// ------------------------------------------------------------------

	ticket, err := fe.CreateTicket(ctx, &newTicketRequest)
	if err != nil {
		log.Printf("Failed to Create Ticket, got %s", err.Error())
	}

	log.Printf("Ticket created successfully for user id %v", ticket.Id)

	// This go routine watches for ticket assignment, once assigned deletes the ticket
	go deleteTicketPostAssignment(ctx, fe, ticket)

	time.Sleep(3 * time.Second)
}

func deleteTicketPostAssignment(ctx context.Context, fe pb.FrontendServiceClient, t *pb.Ticket) {
	// Our code probably should use pubsub events to trigger this
	// Temporary solution to infinitely check for ticket assignment
	for {
		if t.GetAssignment() != nil {
			log.Printf("Ticket %v got assignment %v", t.GetId(), t.GetAssignment())
			break
		}
	}

	_, err := fe.DeleteTicket(ctx, &pb.DeleteTicketRequest{TicketId: t.GetId()})
	if err != nil {
		log.Fatalf("Failed to Delete Ticket %v, got %s", t.GetId(), err.Error())
	}
	log.Printf("Ticket deleted successfully")
}
