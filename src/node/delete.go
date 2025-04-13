package main

import (
	pb "Tapestry/protofiles"
	util "Tapestry/util"
	"context"
	"log"
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

func (n *Node) RTUpdate(ctx context.Context, req *pb.RTUpdateRequest) (*pb.RTUpdateResponse, error) {
	// id := req.ID
	port := int(req.Port)

	replacementID := req.ReplacementID
	replacementPort := int(req.ReplacementPort)
	var found int = 0
	for i, row := range n.RT.Table {
		for j, val := range row {
			if val == port {
				n.RT.Table[i][j] = -1
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

	for i := 0; i < util.DIGITS; i++ {
		id_digit := util.GetDigit(replacementID, i)
		if n.RT.Table[i][id_digit] == -1 {
			n.RT.Table[i][id_digit] = replacementPort

			// connect to update back pointer of replacement node
			conn, to_client, err := GetNodeClient(replacementPort)
			if err != nil {
				log.Panicf("error in connecting (temporary panic): %v", err.Error())
			}
			defer conn.Close()

			// update back pointer
			_, err = to_client.BPUpdate(ctx, &pb.BPUpdateRequest{Id: n.ID, Port: int32(n.Port)})
			if err != nil {
				log.Panicf("error in sending BPUpdate: %v", err.Error())
				return nil, err
			}
			return &pb.RTUpdateResponse{Success: true}, nil

		} else {
			continue
		}
	}

	return &pb.RTUpdateResponse{Success: true}, nil

}

func (n *Node) BPUpdate(ctx context.Context, req *pb.BPUpdateRequest) (*pb.BPUpdateResponse, error) {
	// id := req.Id
	port := int(req.Port)
	n.BP.Set[port] = struct{}{} //inserting into set
	return &pb.BPUpdateResponse{Success: true}, nil
}