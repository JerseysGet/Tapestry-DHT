package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"math/rand"
	"time"
	"context"
	util "Tapestry/util"

	"google.golang.org/grpc"
	pb "Tapestry/Tapestry/protofiles"
)

type Node struct {
	pb.UnimplementedNodeServiceServer
	RT util.RoutingTable
	BP util.BackPointerTable
	ID uint64
}

func (n *Node) Route(ctx context.Context, req *pb.RouteRequest) (*pb.RouteResponse, error) {
	//random implementation of route
	return &pb.RouteResponse{Port: 0}, nil
}
func savePortToFile(port int){

	file, err := os.OpenFile("ports.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close() 

	if _, err := file.WriteString(fmt.Sprintf("%d\n", port)); err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("Port saved to ports.txt")
}

func startSearchForRoot(port int){
	if port == -1 {
		fmt.Println("No root needed")
		return
	}
	fmt.Println("starting search for root using route rpc on port:", port)
	return
}

func main(){

	// get port for search from user
	var port int
	fmt.Print("Enter port to start search: ")
	fmt.Scan(&port)

	// find a free port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatalf("Failed to find a free port: %v", err)
	}
	new_port := listener.Addr().(*net.TCPAddr).Port
	fmt.Println("Server is live on port:", new_port)
	startSearchForRoot(port)
	savePortToFile(new_port)

	// register node server and start listening
	rand.Seed(time.Now().UnixNano())
	nodeID := rand.Uint64()
	var rt util.RoutingTable
	var bp util.BackPointerTable
	node := &Node{
		RT: rt,
		BP: bp,
		ID: nodeID,
	}
	grpcServer := grpc.NewServer()
	pb.RegisterNodeServiceServer(grpcServer, node)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
