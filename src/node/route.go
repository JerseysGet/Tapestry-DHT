package main

import (
	pb "Tapestry/protofiles"
	util "Tapestry/util"
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (n *Node) Route(ctx context.Context, req *pb.RouteRequest) (*pb.RouteResponse, error) {
	id := req.Id
	level := int(req.Level)

	util.Assert(0 <= level && level <= util.DIGITS, "req.level not in bounds")
	if level == util.DIGITS {
		return &pb.RouteResponse{Port: int32(n.Port), Id: n.ID}, nil
	}
	id_digit := util.GetDigit(id, level)
	for d, ct := id_digit, 0; ct < util.RADIX; ct++ {
		// dont LOCK here only reading ?
		to_port := n.RT.Table[level][d]
		if to_port == -1 {
			d = (d + 1) % util.RADIX
			continue
		}

		/* Connect to this port and route on it */
		conn, to_client, err := GetNodeClient(to_port)
		if err != nil {
			// log.Panicf("error in connecting (temporary panic): %v", err.Error())
			d = (d + 1) % util.RADIX
			continue
		}
		defer conn.Close()
		return to_client.Route(ctx, &pb.RouteRequest{Id: id, Level: int32(level + 1)})
	}
	// util.Assert(false, "did not find anyone to route to")
	log.Printf("did not find anyone to route to\n");
	return nil, status.Errorf(codes.NotFound, "no route found for id=%s\n", util.HashToString(id))
}
