package main

import (
	pb "Tapestry/protofiles"
	"context"
)

func (n *Node) BPRemove(ctx context.Context, req *pb.BPRemoveRequest) (*pb.BPRemoveResponse, error) {

	port := int(req.Port)

	if _, exists := n.BP.Set[port]; exists {
		delete(n.BP.Set, port)
	} else {
		return &pb.BPRemoveResponse{Success: false}, nil
	}

	return &pb.BPRemoveResponse{Success: true}, nil
}
