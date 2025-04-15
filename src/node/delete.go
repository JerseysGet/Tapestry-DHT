package main

import (
	pb "Tapestry/protofiles"
	util "Tapestry/util"
	"context"
	"log"
)

func (n *Node) BPRemove(ctx context.Context, req *pb.BPRemoveRequest) (*pb.BPRemoveResponse, error) {

	port := int(req.Port)
	// lock here maybe
	n.BP_lock.RLock()
	_, exists := n.BP.Set[port]
	n.BP_lock.RUnlock()
	if exists {
		n.BP_lock.Lock()
		delete(n.BP.Set, port)
		n.BP_lock.Unlock()
		// PrintRoutingTable()
	} else {
		log.Printf("Port %d not found in back pointer set", port)
		return &pb.BPRemoveResponse{Success: false}, nil
	}

	return &pb.BPRemoveResponse{Success: true}, nil
}

func (n *Node) RTUpdate(ctx context.Context, req *pb.RTUpdateRequest) (*pb.RTUpdateResponse, error) {
	// id := req.ID
	port := int(req.Port)

	replacementID := req.ReplacementID
	replacementPort := int(req.ReplacementPort)
	// lock here maybe
	var found int = 0
	n.RT_lock.Lock()
	for i, row := range n.RT.Table {
		for j, val := range row {
			if val == port {
				n.RT.Table[i][j] = -1
				found = 1
			}
		}
	}

	if found == 0 {
		n.RT_lock.Unlock()
		return &pb.RTUpdateResponse{Success: false}, nil
	}

	if replacementPort == -1 {
		n.RT_lock.Unlock()
		return &pb.RTUpdateResponse{Success: true}, nil
	}

	// update routing table with replacement port
	found = 0
	// lock here maybe
	common_prefix_len := util.CommonPrefixLen(n.ID, replacementID)
	for i := 0; i <= common_prefix_len && i < util.DIGITS; i++ {
		id_digit := util.GetDigit(replacementID, i)
		if n.RT.Table[i][id_digit] == -1 {
			found = 1
			n.RT.Table[i][id_digit] = replacementPort
		}
	}
	n.RT_lock.Unlock()
	// PrintRoutingTable()
	if found == 1 {
		// connect to update back pointer of replacement node
		conn, to_client, err := GetNodeClient(replacementPort)
		if err != nil {
			log.Printf("error in port: %d while connecting to send Back-Pointer update to port : %d, err: %v", n.Port, replacementPort, err.Error())
			return &pb.RTUpdateResponse{Success: false}, nil
		} else {

			// update back pointer
			_, err = to_client.BPUpdate(ctx, &pb.BPUpdateRequest{Id: n.ID, Port: int32(n.Port)})
			if err != nil {
				log.Printf("error in sending BPUpdate: %v", err.Error())
				conn.Close()
				return &pb.RTUpdateResponse{Success: false}, nil
			}
			conn.Close()
		}
	} else {
		log.Printf("No empty slot found in routing table for replacement port %d in node %d\n", replacementPort, n.Port)
	}

	return &pb.RTUpdateResponse{Success: true}, nil

}

func (n *Node) BPUpdate(ctx context.Context, req *pb.BPUpdateRequest) (*pb.BPUpdateResponse, error) {
	// id := req.Id
	port := int(req.Port)
	// lock here maybe
	// log.Println("Inserting into back pointer set", port, n.Port)
	n.BP_lock.Lock()
	n.BP.Set[port] = struct{}{} //inserting into set
	n.BP_lock.Unlock()
	return &pb.BPUpdateResponse{Success: true}, nil
}

func (n *Node) GetID(ctx context.Context, req *pb.GetIDRequest) (*pb.GetIDResponse, error) {
	return &pb.GetIDResponse{ID: n.ID}, nil
}
