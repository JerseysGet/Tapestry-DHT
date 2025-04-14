package main

import (
	pb "Tapestry/protofiles"
	"context"
	"fmt"
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

	fmt.Printf("[PUBLISH] Key '%s' with ID %d stored locally and published to root %d\n",
		key, objectID, rootPort)
	return nil
}

func (n *Node) UnPublish(object Object) error {
	key := object.Name
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

	fmt.Printf("[UNPUBLISH] Key '%s' with ID %d removed locally\n",
		key, objectID)
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

	fmt.Printf("[FIND] Retrieved object '%s' with ID %d from publisher %d\n",
		object.Name, objectID, publisherPort)

	return object, nil
}
