package main

import (
	util "Tapestry/util"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"
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
	Objects           map[uint64]Object           // Object ID -> Object
	Object_Publishers map[uint64]map[int]struct{} // Object ID -> Set of Publisher_Ports
	GrpcServer        *grpc.Server
	Listener          net.Listener
	RT_lock           sync.RWMutex
	Objects_lock      sync.RWMutex
	Publishers_lock   sync.RWMutex
}

func GetNodeClient(port int) (*grpc.ClientConn, pb.NodeServiceClient, error) {
	addr_string := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.NewClient(addr_string, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return conn, pb.NewNodeServiceClient(conn), nil
}

func InitNode(port int, id uint64) *Node {
	addr := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Panicf("could not listen on port %d\n", port)
	}

	actual_port := lis.Addr().(*net.TCPAddr).Port

	ret := &Node{
		RT:                *util.NewRoutingTable(),
		BP:                *util.NewBackPointerTable(),
		ID:                id,
		Port:              actual_port,
		GrpcServer:        grpc.NewServer(),
		Listener:          lis,
		Objects:           make(map[uint64]Object),
		Object_Publishers: make(map[uint64]map[int]struct{}),
	}
	pb.RegisterNodeServiceServer(ret.GrpcServer, ret)
	return ret
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
}

func deleteGracefully(n *Node) {
	var closest_port int
	found := 0
	for i := util.DIGITS - 1; i >= 0; i-- {
		id_digit := util.GetDigit(n.ID, i)
		for j := 0; j < util.RADIX; j++ {
			if uint64(j) == id_digit {
				continue
			}
			if n.RT.Table[i][j] != -1 {
				closest_port = n.RT.Table[i][j]
				found = 1
				break
			}
		}
		if found == 1 {
			break
		}
	}
	var closest_ID uint64
	if found == 0 {
		closest_port = -1
		closest_ID = 0
	} else {
		conn, to_client, err := GetNodeClient(closest_port)
		if err != nil {
			log.Panicf("error in connecting (temporary panic) for GetID: %v", err.Error())
		} else {
			response, err := to_client.GetID(context.Background(), &pb.GetIDRequest{})
			if err != nil {
				log.Panicf("error in GetID: %v", err.Error())
			}
			closest_ID = response.ID
			conn.Close()
		}
	}

	fmt.Printf("closest port found: %d\n", closest_port)
	fmt.Printf("closest ID found: %d\n", util.HashToString(closest_ID))
	// lock here maybe
	// update routing table
	for key_port, _ := range n.BP.Set {
		if key_port == n.Port {
			continue
		}
		conn, to_client, err := GetNodeClient(key_port)
		if err != nil {
			log.Panicf("error in connecting (temporary panic) for RTUpdate: %v", err.Error())
		} else {
			response, err := to_client.RTUpdate(context.Background(), &pb.RTUpdateRequest{ReplacementID: closest_ID, ReplacementPort: int32(closest_port), ID: n.ID, Port: int32(n.Port)})
			if err != nil {
				log.Panicf("error in RTUpdate: %v", err.Error())
			}
			if response.Success {
				fmt.Printf("Routing table updated successfully for port %d\n", key_port)
			} else {
				fmt.Printf("Failed to update routing table for port %d\n", key_port)
			}
			conn.Close()
		}
	}

	// lock here maybe
	// update back pointer table
	for _, row := range n.RT.Table {
		for _, val_port := range row {
			if val_port != n.Port && val_port != -1 {
				conn, to_client, err := GetNodeClient(val_port)
				if err != nil {
					log.Panicf("error in connecting (temporary panic) for BPRemove: %v", err.Error())
				} else {
					response, err := to_client.BPRemove(context.Background(), &pb.BPRemoveRequest{Port: int32(n.Port)})
					if err != nil {
						log.Panicf("error in BPRemove: %v", err.Error())
					}
					if response.Success {
						fmt.Printf("Back pointer table updated successfully for port %d\n", val_port)
					} else {
						fmt.Printf("Failed to remove from Back pointer table for port %d\n", val_port)
					}
					conn.Close()
				}
			}
		}
	}
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
		RT:                rt,
		BP:                bp,
		ID:                nodeID,
		Port:              new_port,
		Objects:           make(map[uint64]Object),
		Object_Publishers: make(map[uint64]map[int]struct{}),
	}
	grpcServer := grpc.NewServer()
	pb.RegisterNodeServiceServer(grpcServer, node)

	// goroutine for republishing
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				for _, obj := range node.Objects {
					err := node.Publish(obj)
					if err != nil {
						fmt.Printf("[RE-PUBLISH ERROR] Object '%s': %v\n", obj.Name, err)
					} else {
						fmt.Printf("[RE-PUBLISH] Object '%s' re-published successfully\n", obj.Name)
					}
				}
			}
		}
	}()

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
