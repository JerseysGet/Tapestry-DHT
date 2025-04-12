package main

import (
	util "Tapestry/util"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	pb "Tapestry/protofiles"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Node struct {
	pb.UnimplementedNodeServiceServer
	RT                util.RoutingTable
	BP                util.BackPointerTable
	ID                uint64
	Port              int
	Objects           map[uint64]Object // Object ID -> Object
	Object_Publishers map[uint64]int    // Object ID -> Publisher_Port
}

func GetNodeClient(port int) (*grpc.ClientConn, pb.NodeServiceClient, error) {
	addr_string := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.NewClient(addr_string, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return conn, pb.NewNodeServiceClient(conn), nil
}

func savePortToFile(port int) {

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

func startSearchForRoot(port int) {
	if port == -1 {
		fmt.Println("No root needed")
		return
	}
	fmt.Println("starting search for root using route rpc on port:", port)
	return
}

func main() {

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
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	nodeID := rng.Uint64()
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
