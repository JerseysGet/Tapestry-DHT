package main

import (
	pb "Tapestry/protofiles"
	"context"
	"fmt"
)

func (n *Node) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	publisherPort := int(req.Port)
	objectID := uint64(req.Object_ID)
	n.Object_Publishers[objectID] = publisherPort
	fmt.Printf("[REGISTER] Received object %d from node %d\n", objectID, publisherPort)
	return &pb.RegisterResponse{}, nil
}

func (n *Node) UnRegister(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	publisherPort := int(req.Port)
	objectID := uint64(req.Object_ID)
	delete(n.Object_Publishers, objectID)
	fmt.Printf("[UNREGISTER] Removed object %d from node %d\n", objectID, publisherPort)
	return &pb.RegisterResponse{}, nil
}

func (n *Node) Lookup(ctx context.Context, req *pb.LookupRequest) (*pb.LookupResponse, error) {
	objectID := uint64(req.Object_ID)
	port, ok := n.Object_Publishers[objectID]
	if !ok {
		return nil, fmt.Errorf("object not found in publishers list")
	}
	return &pb.LookupResponse{
		Port: int32(port),
	}, nil
}

func (n *Node) GetObject(ctx context.Context, req *pb.ObjectRequest) (*pb.ObjectResponse, error) {
	objectID := uint64(req.Object_ID)
	obj, ok := n.Objects[objectID]
	if !ok {
		return nil, fmt.Errorf("object not found locally")
	}
	return &pb.ObjectResponse{
		Name:    obj.Name,
		Content: obj.Content,
	}, nil
}
