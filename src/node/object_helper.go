package main

import (
	pb "Tapestry/protofiles"
	"context"
	"fmt"

	"google.golang.org/grpc"
)

type Object struct {
	Name    string
	Content string
}

type GRPCConnection struct {
	Conn   *grpc.ClientConn
	Client pb.NodeServiceClient
}

func (n *Node) FindRoot(objectID uint64) (int, error) {
	ctx := context.Background()
	resp, err := n.Route(ctx, &pb.RouteRequest{
		Id:    objectID,
		Level: 0,
	})
	if err != nil {
		return 0, fmt.Errorf("routing failed: %v", err)
	}
	return int(resp.Port), nil
}

func (n *Node) ConnectToNode(port int) (*GRPCConnection, error) {
	conn, client, err := GetNodeClient(port)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to node at port %d: %v", port, err)
	}
	return &GRPCConnection{Conn: conn, Client: client}, nil
}

func (n *Node) ConnectToRoot(objectID uint64) (*GRPCConnection, int, error) {
	rootPort, err := n.FindRoot(objectID)
	if err != nil {
		return nil, 0, err
	}
	conn, err := n.ConnectToNode(rootPort)
	if err != nil {
		return nil, 0, err
	}
	return conn, rootPort, nil
}
