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
		Port:      int32(n.Port),
		Object_ID: uint64(objectID),
	})
	if err != nil {
		return fmt.Errorf("failed to register with root: %v", err)
	}
	n.Objects[objectID] = Object{
		Name:    key,
		Content: value,
	}
	fmt.Printf("[PUBLISH] Key '%s' with ID %s stored locally and published to root %d\n",
		key, util.HashToString(objectID), rootPort)

	return nil
}

func (n *Node) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	publisherPort := int(req.Port)
	objectID := uint64(req.Object_ID)
	n.Object_Publishers[objectID] = publisherPort
	fmt.Printf("[REGISTER] Received object %s from node %d\n", util.HashToString(objectID), publisherPort)
	return &pb.RegisterResponse{}, nil
}

func (n *Node) UnPublish(object Object) error {
	key := object.Name
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
	_, err = client.UnRegister(context.Background(), &pb.RegisterRequest{
		Port:      int32(n.Port),
		Object_ID: uint64(objectID),
	})
	if err != nil {
		return fmt.Errorf("failed to register with root: %v", err)
	}
	delete(n.Objects, objectID)
	fmt.Printf("[UNPUBLISH] Key '%s' with ID %s removed locally\n",
		key, util.HashToString(objectID))
	return nil
}

func (n *Node) UnRegister(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	publisherPort := int(req.Port)
	objectID := uint64(req.Object_ID)
	delete(n.Object_Publishers, objectID)
	fmt.Printf("[UNREGISTER] Removed object %s from node %d\n", util.HashToString(objectID), publisherPort)
	return &pb.RegisterResponse{}, nil
}

func (n *Node) FindObject(name string) (Object, error) {
	objectID := util.StringToHash(name)
	ctx := context.Background()
	resp, err := n.Route(ctx, &pb.RouteRequest{
		Id:    objectID,
		Level: 0,
	})
	if err != nil {
		return Object{}, fmt.Errorf("routing failed: %v", err)
	}
	rootPort := int(resp.Port)

	connRoot, clientRoot, err := GetNodeClient(rootPort)
	if err != nil {
		return Object{}, fmt.Errorf("failed to connect to root node: %v", err)
	}
	defer connRoot.Close()
	lookupResp, err := clientRoot.Lookup(ctx, &pb.LookupRequest{
		Object_ID: uint64(objectID),
	})
	if err != nil {
		return Object{}, fmt.Errorf("lookup failed: %v", err)
	}
	publisherPort := int(lookupResp.Port)

	connPub, clientPub, err := GetNodeClient(publisherPort)
	if err != nil {
		return Object{}, fmt.Errorf("failed to connect to publisher node: %v", err)
	}
	defer connPub.Close()
	objResp, err := clientPub.GetObject(ctx, &pb.ObjectRequest{
		Object_ID: uint64(objectID),
	})
	if err != nil {
		return Object{}, fmt.Errorf("failed to get object from publisher: %v", err)
	}

	object := Object{
		Name:    objResp.Name,
		Content: objResp.Content,
	}
	fmt.Printf("[FIND] Retrieved object '%s' with ID %s from publisher %d\n",
		object.Name, util.HashToString(objectID), publisherPort)

	return object, nil
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
