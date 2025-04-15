package main

import (
	pb "Tapestry/protofiles"
	"context"
	"fmt"
	"log"
	"strings"
)

func (n *Node) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	publisherPort := int(req.Port)
	objectID := uint64(req.Object_ID)

	n.Publishers_lock.Lock()
	defer n.Publishers_lock.Unlock()

	if _, ok := n.Object_Publishers[objectID]; !ok {
		n.Object_Publishers[objectID] = make(map[int]struct{})
	}
	n.Object_Publishers[objectID][publisherPort] = struct{}{}

	log.Printf("[REGISTER] Received object %d from node %d\n", objectID, publisherPort)
	return &pb.RegisterResponse{}, nil
}

func (n *Node) UnRegister(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	publisherPort := int(req.Port)
	objectID := uint64(req.Object_ID)

	n.Publishers_lock.Lock()
	portSet, ok := n.Object_Publishers[objectID]
	if ok {
		for port := range portSet {
			if port == publisherPort {
				continue
			}
			conn, client, err := GetNodeClient(port)
			if err != nil {
				log.Printf("Failed to connect to node %d for RemoveObject: %v", port, err)
				continue
			}
			defer conn.Close()
			_, err = client.RemoveObject(context.Background(), &pb.RemoveObjectRequest{
				Object_ID: objectID,
			})
			if err != nil && !strings.Contains(err.Error(), "Unavailable") {
				log.Printf("RemoveObject failed on node %d: %v", port, err)
			} else {
				log.Printf("[UNREGISTER] Informed node %d to remove object %d\n", port, objectID)
			}
		}
		delete(n.Object_Publishers, objectID)
	}
	n.Publishers_lock.Unlock()

	log.Printf("[UNREGISTER] Removed object %d from node %d\n", objectID, publisherPort)
	return &pb.RegisterResponse{}, nil
}

func (n *Node) RemoveObject(ctx context.Context, req *pb.RemoveObjectRequest) (*pb.RemoveObjectResponse, error) {
	objectID := uint64(req.Object_ID)

	n.Objects_lock.Lock()
	delete(n.Objects, objectID)
	n.Objects_lock.Unlock()

	log.Printf("[REMOVE OBJECT] Object %d removed from node %d\n", objectID, n.Port)
	return &pb.RemoveObjectResponse{}, nil
}

func (n *Node) Lookup(ctx context.Context, req *pb.LookupRequest) (*pb.LookupResponse, error) {
	objectID := uint64(req.Object_ID)

	n.Publishers_lock.Lock()
	defer n.Publishers_lock.Unlock()

	portSet, ok := n.Object_Publishers[objectID]
	if !ok || len(portSet) == 0 {
		return &pb.LookupResponse{Port: -1}, nil
	}
	fmt.Printf("Portset %v\n", portSet)
	for port := range portSet {
		// fmt.Printf("Trying to contact %d\n", port)
		conn, clnt, err := GetNodeClient(port)
		if err != nil {
			// fmt.Printf("Couldnt contact %d\n", port)
			delete(portSet, port)
			continue
		}

		_, err = clnt.Ping(context.Background(), &pb.Nothing{})
		if err != nil {
			// fmt.Printf("Couldnt contact %d\n", port)
			delete(portSet, port)
			continue
		}

		fmt.Printf("Contacted %d\n", port)
		conn.Close()
		return &pb.LookupResponse{Port: int32(port)}, nil
	}

	delete(n.Object_Publishers, objectID)
	return &pb.LookupResponse{Port: -1}, nil
}

func (n *Node) GetObject(ctx context.Context, req *pb.ObjectRequest) (*pb.ObjectResponse, error) {
	objectID := uint64(req.Object_ID)

	n.Objects_lock.RLock()
	obj, ok := n.Objects[objectID]
	n.Objects_lock.RUnlock()

	if !ok {
		return nil, fmt.Errorf("object not found locally")
	}
	return &pb.ObjectResponse{
		Name:    obj.Name,
		Content: obj.Content,
	}, nil
}

func (n *Node) StoreObject(ctx context.Context, obj *pb.Object) (*pb.Ack, error) {
	object := Object{
		Name:    obj.GetName(),
		Content: obj.GetContent(),
	}
	objectID := StringToUint64(object.Name)

	n.Objects_lock.Lock()
	n.Objects[objectID] = object
	n.Objects_lock.Unlock()

	return &pb.Ack{Success: true}, nil
}
