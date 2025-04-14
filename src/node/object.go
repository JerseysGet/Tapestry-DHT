package main

import (
	pb "Tapestry/protofiles"
	"context"
	"fmt"
	"log"
	"math/rand"
)

func (n *Node) Publish(object Object) error {
	key := object.Name
	value := object.Content
	objectID := StringToUint64(key)

	conn, rootPort, err := n.ConnectToRoot(objectID)
	if err != nil {
		return err
	}
	defer conn.Conn.Close()

	_, err = conn.Client.Register(context.Background(), &pb.RegisterRequest{
		Port:      int32(n.Port),
		Object_ID: uint64(objectID),
	})
	if err != nil {
		return fmt.Errorf("failed to register with root: %v", err)
	}

	n.Objects_lock.Lock()
	n.Objects[objectID] = Object{
		Name:    key,
		Content: value,
	}
	n.Objects_lock.Unlock()

	fmt.Printf("[PUBLISH] Key '%s' with ID %d stored locally and published to root %d\n", key, objectID, rootPort)
	return nil
}

func (n *Node) UnPublish(key string) error {
	objectID := StringToUint64(key)

	conn, _, err := n.ConnectToRoot(objectID)
	if err != nil {
		return err
	}
	defer conn.Conn.Close()

	_, err = conn.Client.UnRegister(context.Background(), &pb.RegisterRequest{
		Port:      int32(n.Port),
		Object_ID: uint64(objectID),
	})
	if err != nil {
		return fmt.Errorf("failed to unregister with root: %v", err)
	}

	n.Objects_lock.Lock()
	delete(n.Objects, objectID)
	n.Objects_lock.Unlock()

	// fmt.Printf("[UNPUBLISH] Key '%s' with ID %d removed locally\n",key, objectID)
	return nil
}

func (n *Node) FindObject(name string) (Object, error) {
	objectID := StringToUint64(name)

	rootConn, _, err := n.ConnectToRoot(objectID)
	if err != nil {
		return Object{}, err
	}
	defer rootConn.Conn.Close()

	ctx := context.Background()
	lookupResp, err := rootConn.Client.Lookup(ctx, &pb.LookupRequest{
		Object_ID: uint64(objectID),
	})
	if err != nil {
		return Object{}, fmt.Errorf("lookup failed: %v", err)
	}
	publisherPort := int(lookupResp.Port)

	pubConn, err := n.ConnectToNode(publisherPort)
	if err != nil {
		return Object{}, err
	}
	defer pubConn.Conn.Close()

	objResp, err := pubConn.Client.GetObject(ctx, &pb.ObjectRequest{
		Object_ID: uint64(objectID),
	})
	if err != nil {
		return Object{}, fmt.Errorf("failed to get object from publisher: %v", err)
	}

	object := Object{
		Name:    objResp.Name,
		Content: objResp.Content,
	}

	// fmt.Printf("[FIND] Retrieved object '%s' with ID %d from publisher %d\n", object.Name, objectID, publisherPort)

	return object, nil
}

func (n *Node) AddObject(obj Object) error {
	seen := make(map[int]struct{})
	maxCandidates := 0

	n.RT_lock.RLock()
	for level := 0; level < len(n.RT.Table); level++ {
		for digit := 0; digit < len(n.RT.Table[level]); digit++ {
			port := n.RT.Table[level][digit]
			if port != -1 && port != n.Port {
				maxCandidates++
			}
		}
	}
	n.RT_lock.RUnlock()

	if maxCandidates == 0 {
		// log.Println("[ADD OBJECT] No other valid nodes found to replicate object")
	} else {
		added := 0
		attempts := 0
		for added < 2 && attempts < maxCandidates*2 {
			attempts++

			n.RT_lock.RLock()
			level := rand.Intn(len(n.RT.Table))
			digit := rand.Intn(len(n.RT.Table[level]))
			port := n.RT.Table[level][digit]
			n.RT_lock.RUnlock()

			if port == -1 || port == n.Port {
				continue
			}
			if _, exists := seen[port]; exists {
				continue
			}
			seen[port] = struct{}{}
			added++

			go func(port int) {
				conn, client, err := GetNodeClient(port)
				if err != nil {
					log.Printf("Could not connect to node at port %d: %v", port, err)
					return
				}
				defer conn.Close()

				pbObj := &pb.Object{
					Name:    obj.Name,
					Content: obj.Content,
				}

				_, err = client.StoreObject(context.Background(), pbObj)
				if err != nil {
					log.Printf("Error storing object on port %d: %v", port, err)
				}
			}(port)
		}

		if added < 2 {
			// log.Printf("[ADD OBJECT] Only %d replicas could be added (less than desired 2)", added)
		}
	}

	objectID := StringToUint64(obj.Name)
	n.Objects_lock.Lock()
	n.Objects[objectID] = obj
	n.Objects_lock.Unlock()

	return nil
}
