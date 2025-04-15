#import "../header.typ" as h
#show: h.my-template
#h.title("Decentralised overlay network - Tapestry")
#h.subtitle("Project Report - Team 26")
#h.author("Rohan Sridhar (2022101042) Mohammed Faisal (2022101101) Shreyansh (2022111002)")

= Project Title
Implementation of a Decentralized Overlay Network Inspired by Tapestry

= Problem Statement
This project implements a decentralized overlay network inspired by Tapestry, enabling scalable, fault-tolerant routing with prefix-based matching. It supports efficient message routing, dynamic node membership, and resilience to node failures for fast lookups in large-scale distributed systems.

= Framework and Technologies
- *Programming Language:* Go (Golang)
- *Communication Protocol:* gRPC 
- *Data Structures:* Prefix-based routing tables and Back pointers
- *Hashing Mechanism:* FNV-A1 hash
- *Storage (Resources):* In-memory storage for node states

== Reasoning Behind Technology Choices
- *Go (Golang) & gRPC*: gRPC provides high-performance,  remote procedure calls (RPCs) with built-in support for error handling and serialization using Protocol Buffers. All of us have experience with gRPC in Go from the Assignment.

- *Prefix-Based Routing Tables*: The original Tapestry paper implements a prefix based routing system, using SHA-1, so we are implementing similar to that.

= Project Functionalities
== *Node Insertion:*
- Nodes can join the network and establish connections with existing nodes ( Populates routing tables and back pointers ).
- Each node generates a unique random 64-bit ID using the FNV-A1 hash function.
== *Routing:*
- Routing method is used to find the root node corresponding to the given key ( which can be either Node ID or Object hash ).
- The routing method uses the prefix-based routing algorithm to find the node responsible for the given key.
- The returning result is the port of the node with maximum common prefix length with the key.
== *Node Deletion:*
- Nodes can leave the network gracefully.
- The routing tables and back pointers are updated accordingly.
- The exiting node won't be accessible after exit.
== *Add Object:*
- Nodes can add objects to the network
- They take object inputs from the users
- The objects are key-value pairs
== *Object Publish:*
- Nodes can publish objects to the network
- Any node can access the object value using their keys after they are published
== *Object Unpublish:*
- Nodes can unpublish objects from the network
- The object is removed from the network and no node would be able to access object after This
== *Find Object:*
- Nodes can find objects in the network
- They can find the root node for the given object key and access the value by asking object's node-port from root node
== *Fault Tolerance:*
- The system can handle node failures and reconfigure routing tables.
- Even after a node goes down unexpectedly, the system can still function.
== *Redundancy:*
- The system maintains redundancy by keeping multiple copies of objects, so that even after a node goes down, it's objects are accessible from redundent resources

= Implementation Details 
== *Node Insertion:*
== *Routing:*
== *Node Deletion:*
- For graceful deletion, a node first identifies the closest node to its own ID, which then serves as its replacement during the deletion process.
- The replacement node updates routing tables and fills any gaps left by the departing node, ensuring continuity and network integrity during deletion.
- Node deletion involves three key steps: *RTUpdate*, to update routing tables; *BPUpdate*, to update backpointers; and *BPRemove*, to remove the departing node from others' backpointers.
- RTUpdate: Updates the routing tables of all nodes (excluding the departing node) by removing references to the exiting node and filling any gaps with the replacement node's information.
- BPUpdate: Invoked within the RTUpdate process, this step requests the replacement node to update its backpointer table when other nodes add it to their routing tables.
- BPRemove: The final step, which removes the exiting node from the backpointer tables of all other nodes in the network.
- Once all routing and backpointer tables are updated across the network, the exiting node gracefully leaves the system.
== *Add Object:*
- Accepts an object as input from the user and stores it in the node’s local `Objects` map.
- Ensures redundancy by invoking the `StoreObject()` RPC on up to two other nodes, selected by scanning the routing table. These nodes then replicate the object in their respective local maps.

== *Object Publish:*
- The `Publish()` function is periodically invoked every 5 seconds in a separate goroutine.
- It first identifies the root node using `FindRoot()`, which internally calls `Route()`.
- The function then calls the `Register()` RPC on the root node to register itself as a publisher of the object.
- This repeated invocation provides fault tolerance, ensuring that in the event of a failure, a new root is automatically assigned and updated.

== *Object Unpublish:*
- Removes the object from the node’s local `Objects` map.
- Sends an `Unregister()` RPC to the root node, which in turn instructs all other publishers of the object to remove it from their local storage as well.
- This ensures consistency across the system while preserving the desired redundancy.

== *Find Object:*
- Retrieves the object specified by the user by contacting one of its active publishers.
- Calls the `LookUp()` RPC on the root node to obtain the port number of a live publisher for the requested object.
- It then calls the `GetObject()` RPC on the selected publisher to fetch the object.
- If the object is successfully found, it is returned. Otherwise, a message indicating the absence of the object is displayed.

= Performance Test Results

