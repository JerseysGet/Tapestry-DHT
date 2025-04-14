package main

import (
	pb "Tapestry/protofiles"
	util "Tapestry/util"
	"context"
	"fmt"
	"log"
)

func (n *Node) InformHoleMulticast(ctx context.Context, req *pb.MulticastRequest) (*pb.MulticastResponse, error) {
	level := int(req.Level)
	original_level := int(req.OriginalLevel)
	new_port := int(req.NewPort)
	new_id := req.NewID
	util.Assert(0 <= level, "negative level")

	if level < util.DIGITS {
		for d := range util.RADIX {
			// dont LOCK only reading here?
			n.RT_lock.RLock()
			to_port := n.RT.Table[level][d]
			n.RT_lock.RUnlock()
			if to_port == -1 {
				continue
			}
			if to_port == new_port {
				continue
			}
			conn, to_client, err := GetNodeClient(to_port)
			if err != nil {
				log.Panicf("error in connecting (temporary panic): %v", err.Error())
			}
			defer conn.Close()
			to_client.InformHoleMulticast(ctx, &pb.MulticastRequest{NewPort: req.NewPort, NewID: req.NewID, OriginalLevel: req.OriginalLevel, Level: req.Level + 1})
		}
	}
	// going to do n.RT[level][getdigit(new_id, level)] = new_port
	// LOCK HERE maybe
	digit := util.GetDigit(new_id, original_level)
	n.RT_lock.Lock()
	n.RT.Table[original_level][digit] = new_port
	n.RT_lock.Unlock()
	PrintRoutingTable()
	conn, new_client, err := GetNodeClient(new_port)
	if err != nil {
		log.Panicf("error in connecting (temporary panic): %v", err.Error())
	}
	defer conn.Close()
	new_client.BPUpdate(context.Background(), &pb.BPUpdateRequest{Id: n.ID, Port: int32(n.Port)})
	return &pb.MulticastResponse{Status: 0}, nil
}

func (n *Node) RTCopy(ctx context.Context, req *pb.Nothing) (*pb.RTCopyReponse, error) {
	n.RT_lock.RLock()
	rt_copy := n.RT.Table
	n.RT_lock.RUnlock()
	ret := &pb.RTCopyReponse{
		Rows: int32(util.DIGITS),
		Cols: int32(util.RADIX),
		Data: util.FlattenMatrix(rt_copy), // don't LOCK here only reading ?
	}
	return ret, nil
}

func (n *Node) Insert(BootstrapPort int) error {
	n.RT_lock.Lock()
	for level := range util.DIGITS {
		n.RT.Table[level][util.GetDigit(n.ID, level)] = n.Port
	}
	n.RT_lock.Unlock()
	if BootstrapPort == 0 {
		return nil
	}
	conn, boot_client, err := GetNodeClient(BootstrapPort)
	if err != nil {
		fmt.Print(err.Error())
		return err
	}
	resp, err := boot_client.Route(context.Background(), &pb.RouteRequest{Id: n.ID, Level: 0})
	if err != nil {
		return err
	}
	conn.Close()
	root_port := int(resp.Port)
	root_id := resp.Id
	// copy RT of root_port
	root_conn, root_client, err := GetNodeClient(root_port)
	if err != nil {
		log.Panicf("error in connecting (temporary panic): %v", err.Error())
	}
	root_resp, err := root_client.RTCopy(context.Background(), &pb.Nothing{})
	if err != nil {
		return err
	}
	common_len := util.CommonPrefixLen(root_id, n.ID)
	root_client.InformHoleMulticast(context.Background(), &pb.MulticastRequest{NewPort: int32(n.Port), NewID: n.ID, OriginalLevel: int32(common_len), Level: int32(common_len)})
	root_conn.Close()

	rt_copy := util.UnflattenMatrix(root_resp.Data, int(root_resp.Rows), int(root_resp.Cols))

	for level := common_len + 1; level < util.DIGITS; level++ {
		for d := range util.RADIX {
			rt_copy[level][d] = -1
		}
	}

	for level := range util.DIGITS {
		rt_copy[level][util.GetDigit(n.ID, level)] = n.Port
	}

	for level := range util.DIGITS {
		for d := range util.RADIX {
			port := rt_copy[level][d]
			if port == -1 {
				continue
			}
			if port == n.Port {
				continue
			}
			conn, client, err := GetNodeClient(port)
			if err != nil {
				log.Panicf("error in connecting (temporary panic): %v", err.Error())
			}
			client.BPUpdate(context.Background(), &pb.BPUpdateRequest{Id: n.ID, Port: int32(n.Port)})
			conn.Close()
		}
	}

	// LOCK HERE ?
	n.RT_lock.Lock()
	n.RT.Table = rt_copy
	n.RT_lock.Unlock()
	PrintRoutingTable()
	return nil
}
