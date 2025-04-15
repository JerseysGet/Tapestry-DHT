package main

import (
	util "Tapestry/util"
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
	"path/filepath"
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
	BP_lock           sync.RWMutex
	Objects_lock      sync.RWMutex
	Publishers_lock   sync.RWMutex
}

func GetNodeClient(port int) (*grpc.ClientConn, pb.NodeServiceClient, error) {
	addr_string := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.NewClient(addr_string, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("could not connect to port %d\n", port)
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

func setupLogger(port int) {
	logDir := filepath.Join(".", "logs")
	logFilePath := filepath.Join(logDir, fmt.Sprintf("%d.log", port))
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	log.SetOutput(file)
}

func PrintRoutingTable() {
	fmt.Printf("ID= %d (%s)\n", Self.ID, util.HashToString(Self.ID))
	for level := 0; level < util.DIGITS; level++ {
		fmt.Printf("  Level %d: ", level)
		for digit := 0; digit < util.RADIX; digit++ {
			if Self.RT.Table[level][digit] != -1 {
				fmt.Printf("%d ", Self.RT.Table[level][digit])
			} else {
				fmt.Print(". ")
			}
		}
		fmt.Println()
	}
	fmt.Printf("Back Pointer Table for port: %d ID:%s\n", Self.Port, util.HashToString(Self.ID))
	for port := range Self.BP.Set {
		fmt.Printf("%d ", port)
	}
	fmt.Println()
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
	var closest_ID uint64
	found := 0
	n.RT_lock.Lock()
	for i := util.DIGITS - 1; i >= 0; i-- {
		id_digit := util.GetDigit(n.ID, i)
		for j := 0; j < util.RADIX; j++ {
			if uint64(j) == id_digit {
				continue
			}
			if n.RT.Table[i][j] != -1 {
				closest_port = n.RT.Table[i][j]
				found = 1
				conn, to_client, err := GetNodeClient(closest_port)
				if err != nil {
					log.Printf("error in connecting for GetID: %v", err.Error())
					found = 0
					continue
				} else {
					response, err := to_client.GetID(context.Background(), &pb.GetIDRequest{})
					if err != nil {
						log.Printf("error in GetID: %v", err.Error())
						found = 0
						conn.Close()
						continue
					}
					closest_ID = response.ID
					conn.Close()
				}
				break
			}
		}
		if found == 1 {
			break
		}
	}
	
	if found == 0 {
		closest_port = -1
		closest_ID = 0
	} 
	n.RT_lock.Unlock()

	fmt.Printf("closest port found: %d\n", closest_port)
	fmt.Printf("closest ID found: %s\n", util.HashToString(closest_ID))
	// lock here maybe
	// update routing table
	n.BP_lock.Lock()
	for key_port, _ := range n.BP.Set {
		if key_port == n.Port {
			continue
		}
		conn, to_client, err := GetNodeClient(key_port)
		if err != nil {
			log.Printf("error in connecting for RTUpdate: %v", err.Error())
			continue
		} else {
			response, err := to_client.RTUpdate(context.Background(), &pb.RTUpdateRequest{ReplacementID: closest_ID, ReplacementPort: int32(closest_port), ID: n.ID, Port: int32(n.Port)})
			if err != nil {
				log.Printf("error in RTUpdate: %v", err.Error())
				conn.Close()
				continue
			}
			if response.Success {
				log.Printf("Routing table updated successfully for port %d\n", key_port)
			} else {
				log.Printf("Failed to update routing table for port %d\n", key_port)
			}
			conn.Close()
		}
	}

	n.BP_lock.Unlock()

	// lock here maybe
	// update back pointer table
	n.RT_lock.Lock()
	for _, row := range n.RT.Table {
		for _, val_port := range row {
			if val_port != n.Port && val_port != -1 {
				conn, to_client, err := GetNodeClient(val_port)
				if err != nil {
					log.Printf("error in connecting for BPRemove: %v", err.Error())
					continue
				} else {
					response, err := to_client.BPRemove(context.Background(), &pb.BPRemoveRequest{Port: int32(n.Port)})
					if err != nil {
						log.Printf("error in BPRemove: %v", err.Error())
						conn.Close()
						continue
					}
					if response.Success {
						log.Printf("Back pointer table updated successfully for port %d\n", val_port)
					} else {
						log.Printf("Failed to remove from Back pointer table for port %d\n", val_port)
					}
					conn.Close()
				}
			}
		}
	}
	n.RT_lock.Unlock()
}

var Self *Node
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func RepublishObjects() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			for _, obj := range Self.Objects {
				err := Self.Publish(obj)
				if err != nil {
					fmt.Printf("[RE-PUBLISH ERROR] Object '%s': %v\n", obj.Name, err)
				} else {
					// fmt.Printf("[RE-PUBLISH] Object '%s' re-published successfully\n", obj.Name)
				}
			}
		}
	}
}

func TakeInput() {
	var port int
	var id_str string
	fmt.Print("Enter port (0 for random): ")
	fmt.Scan(&port)
	fmt.Print("Enter ID (0 for random): ")
	fmt.Scan(&id_str)
	var id uint64
	if id_str == "0" {
		id = rng.Uint64()
	} else {
		id = util.StringToHash(util.PadLeft32((id_str)))
	}
	Self = InitNode(port, id)
	fmt.Printf("Port=%d, ID=%s\n", Self.Port, util.HashToString(Self.ID))
	go func() {
		if err := Self.GrpcServer.Serve(Self.Listener); err != nil {
			log.Panic("could not serve\n")
		}
	}()
}

func main() {
	TakeInput()
	var boot_port int
	fmt.Print("Enter bootstrap port (0 for empty network): ")
	fmt.Scan(&boot_port)
	setupLogger(Self.Port)
	err := Self.Insert(boot_port)
	// PrintRoutingTable()
	if err != nil {
		log.Print(err.Error())
		log.Println("could not insert")
		os.Exit(1)
	}
	log.Println("Inserted successfully")

	go RepublishObjects()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)
	input := make(chan string)
	scanner := bufio.NewScanner(os.Stdin)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			input <-line
		}
	}()

	for {
		fmt.Println("\nChoose an option:")
		fmt.Println("[1] Publish")
		fmt.Println("[2] Find Object")
		fmt.Println("[3] Unpublish")
		fmt.Println("[4] Exit")

		// var choice int
		fmt.Print("Enter choice: ")
		// fmt.Scan(&choice)

		select {
		case sig := <-sigs:
			// time.Sleep(1000 * time.Millisecond)
			fmt.Printf("\nReceived signal: %s. Exiting...\n", sig)
			fmt.Println("Exiting.")
			deleteGracefully(Self)
			time.Sleep(500 * time.Millisecond)
			Self.GrpcServer.GracefulStop()
			fmt.Println("gRPC server stopped.")
			return
		case line := <-input:
			line = strings.TrimSpace(line)
			switch line {
			case "1":
				var objectName, objectContent string
				fmt.Print("Enter object name: ")
				// fmt.Scanf("%s", &objectName) // Read the whole line for object name
				objectName = <-input
				// Consume any extra newline characters in the input buffer
				// fmt.Scanln() is a safe way to clear the buffer after scanning strings
	
				fmt.Print("Enter object content: ")
				// fmt.Scanf("%s", &objectContent) // Read the whole line for object content
				objectContent = <-input
				obj := Object{
					Name:    objectName,
					Content: objectContent,
				}
				err := Self.AddObject(obj)
				if err != nil {
					fmt.Printf("Error publishing object: %v\n", err)
				} else {
					fmt.Println("Object successfully added and published!")
				}
			case "2":
				fmt.Println("Finding Object...")
				var objectName string
				fmt.Print("Enter object name: ")
				// fmt.Scanln(&objectName) // Read the whole line for object name
				objectName = <-input
				object, err := Self.FindObject(objectName)
				if err != nil {
					fmt.Printf("Error finding object: %v\n", err)
				} else {
					fmt.Printf("Object found! Name: %s, Content: %s\n", object.Name, object.Content)
				}
			case "3":
				var objectName string
				fmt.Print("Enter object name: ")
				// fmt.Scanln(&objectName) // Read the whole line for object name
				objectName = <-input
				err := Self.UnPublish(objectName)
				if err != nil {
					fmt.Printf("Error unpublishing object: %v\n", err)
				} else {
					fmt.Println("Object successfully unpublished!")
				}
			case "4":
				fmt.Println("Exiting.")
				deleteGracefully(Self)
				time.Sleep(500 * time.Millisecond)
				Self.GrpcServer.GracefulStop()
				fmt.Println("gRPC server stopped.")
				return
			default:
				fmt.Println("Invalid choice. Try again.")
			}
		}

		// switch choice {
		// case 1:
		// 	var objectName, objectContent string
		// 	fmt.Print("Enter object name: ")
		// 	fmt.Scanf("%s", &objectName) // Read the whole line for object name

		// 	// Consume any extra newline characters in the input buffer
		// 	// fmt.Scanln() is a safe way to clear the buffer after scanning strings

		// 	fmt.Print("Enter object content: ")
		// 	fmt.Scanf("%s", &objectContent) // Read the whole line for object content

		// 	obj := Object{
		// 		Name:    objectName,
		// 		Content: objectContent,
		// 	}
		// 	err := Self.AddObject(obj)
		// 	if err != nil {
		// 		fmt.Printf("Error publishing object: %v\n", err)
		// 	} else {
		// 		fmt.Println("Object successfully added and published!")
		// 	}
		// case 2:
		// 	fmt.Println("Finding Object...")
		// 	var objectName string
		// 	fmt.Print("Enter object name: ")
		// 	fmt.Scanln(&objectName) // Read the whole line for object name

		// 	object, err := Self.FindObject(objectName)
		// 	if err != nil {
		// 		fmt.Printf("Error finding object: %v\n", err)
		// 	} else {
		// 		fmt.Printf("Object found! Name: %s, Content: %s\n", object.Name, object.Content)
		// 	}
		// case 3:
		// 	var objectName string
		// 	fmt.Print("Enter object name: ")
		// 	fmt.Scanln(&objectName) // Read the whole line for object name

		// 	err := Self.UnPublish(objectName)
		// 	if err != nil {
		// 		fmt.Printf("Error unpublishing object: %v\n", err)
		// 	} else {
		// 		fmt.Println("Object successfully unpublished!")
		// 	}
		// case 4:
		// 	fmt.Println("Exiting.")
		// 	deleteGracefully(Self)
		// 	time.Sleep(500 * time.Millisecond)
		// 	Self.GrpcServer.GracefulStop()
		// 	fmt.Println("gRPC server stopped.")
		// 	return
		// default:
		// 	fmt.Println("Invalid choice. Try again.")
		// }
	}
}
