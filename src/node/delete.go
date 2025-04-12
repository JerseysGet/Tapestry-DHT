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

func (n *node) RTUpdate(ctx context.Context, req *pb.RTUpdateRequest) (*pb.RTUpdateResponse, error) {
	id := req.ID
	port := int(req.Port)

	replacementID := req.ReplacementID
	replacementPort := int(req.ReplacementPort)
	var found int = 0
	for i, row := range n.RT.Table {
		for j, val := range row {
			if val == port {
				n.RT.Table[i][j] = replacementPort
				found = 1
				break
			}
			if found == 1 {
				break
			}
		}
	}
	if found == 0 {
		return &pb.RTUpdateResponse{Success: false}, nil
	}
	return &pb.RTUpdateResponse{Success: true}, nil

}