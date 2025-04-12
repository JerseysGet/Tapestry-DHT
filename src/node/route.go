package main

import (
	pb "Tapestry/protofiles"
	util "Tapestry/util"
	"context"
	"log"
)


func (n *Node) Route(ctx context.Context, req *pb.RouteRequest) (*pb.RouteResponse, error) {
	id := req.Id
	level := int(req.Level)

	util.Assert(0 <= level && level < util.DIGITS, "req.level not in bounds")
	if level + 1 == util.DIGITS {
		return &pb.RouteResponse{Port: int32(n.Port), Id: n.ID}, nil
	}
	id_digit := util.GetDigit(id, level)
	for d, ct := id_digit, 0; ct < util.RADIX; ct++ {
		to_port := n.RT.Table[level][d]
		if to_port == -1 { 
			d = (d + 1) % util.RADIX  
			continue
		}
		
		/* Connect to this port and route on it */
		conn, to_client, err := GetNodeClient(to_port)
		if err != nil {
			log.Panicf("error in connecting (temporary panic): %v", err.Error())
		}
		defer conn.Close()
		return to_client.Route(ctx, &pb.RouteRequest{Id: id, Level: int32(level + 1)})
	}
	util.Assert(false, "did not find anyone to route to")
	return nil, nil
}

