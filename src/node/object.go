package main

import (
	pb "Tapestry/protofiles"
	util "Tapestry/util"
	"context"
	"fmt"
)

type Object struct {
	Name    string
	Content string
}

func (n *Node) Publish(object Object) error {
	key := object.Name
	value := object.Content
	objectID := util.StringToHash(key)
	ctx := context.Background()
	resp, err := n.Route(ctx, &pb.RouteRequest{
		Id:    objectID,
		Level: 0,
	})
	if err != nil {
		return fmt.Errorf("routing failed: %v", err)
	}
	rootPort := int(resp.Port)
	conn, client, err := GetNodeClient(rootPort)
	if err != nil {
		return fmt.Errorf("failed to connect to root node: %v", err)
	}
	defer conn.Close()
	_, err = client.Register(context.Background(), &pb.RegisterRequest{
		Port: int32(n.Port),
	})
	if err != nil {
		return fmt.Errorf("failed to register with root: %v", err)
	}

	n.Objects[objectID] = Object{
		Name:    key,
		Content: value,
	}
	n.Object_roots[objectID] = rootPort

	fmt.Printf("[PUBLISH] Key '%s' with ID %s stored locally and published to root %d\n",
		key, util.HashToString(objectID), rootPort)

	return nil
}

func (n *Node) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	objectID := req.Id
	publisherPort := int(req.Port)

	// Log the replica
	fmt.Printf("[REGISTER] Received object %s from node %d\n", util.HashToString(objectID), publisherPort)

	// You could store a list of replicas per objectID if needed
	// For now, just acknowledge receipt
	return &pb.RegisterResponse{}, nil
}
