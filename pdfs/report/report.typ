#import "../header.typ" as h
#show: h.my-template
#h.title("Decentralised overlay network - Tapestry")
#h.subtitle("Project Report - Team 26")
#h.long_date(15, 4, 2025)
#linebreak()
#h.author("Rohan Sridhar (2022101042) Mohammed Faisal (2022101101) Shreyansh (2022111002)")

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
- Nodes can join the network and establish connections with existing nodes (Populates routing tables and back pointers).
- Each node generates a unique random 64-bit ID using the FNV-A1 hash function.
== *Routing:*
- Routing method is used to find the root node corresponding to a given key (which can be either Node ID / Object hash) in $O(1)$ hops.
- The routing method uses the prefix-based routing algorithm to find the node responsible for the given key.
- The result is the port of the node with maximum common prefix length with the key.
== *Node Deletion:*
- Nodes can leave the network gracefully.
- The routing tables and back pointers are updated accordingly in $O(log^2 n)$ messages.
- The exiting node won't be accessible after exit.
== *Add Object:*
- Objects are (key, value) pairs (like a *distributed hash table*) that are stored in the network.
- Nodes can add objects to the network, which can then be located from anywhere.
== *Object Publish/Unpublish/Find:*
- Nodes can publish objects to the network
- Any node can access the object value using their keys after they are published
- An object can be unpublished from anywhere in the network

== *Fault Tolerance:*
- The system can handle node failures and reconfigure routing tables.
- Even after a node goes down unexpectedly, the system can still function.

== *Redundancy:*
- The system maintains redundancy by keeping multiple copies of objects, so that even after a node goes down, its objects are accessible from redundant resources.
#pagebreak()
= Implementation Details 
== Radix, Hash length considerations
- Original Tapestry implementation uses base 16 with a 160-bit hash, which gives 40 digits.
- To simplify the implementation, a 64-bit hash is  used with base 4, to give 32 digits.
== *Node Insertion:*
- New nodes are randomly assigned 64-bit IDs, and are inserted with the help of a bootstrap node.
- The bootstrap node routes the new ID to a (unique) root, that has the *longest common prefix* with the new ID.
- The Routing Table of the new node is obtained by copying the routing table of the node for levels < longest common prefix.
- Higher levels are filled with a *multicast* operation.
- Random assignment of IDs gives $O(log n)$ nodes that are contacted in the multicast (refer #link(<appendix>, text("Appendix"
, fill: blue))).


== *Routing:*
- A prefix-based routing algorithm is used, very similar to a search on a Trie. The Routing tables maintained make up a *Distributed Trie*.
- Since the IDs are 64-bit hashes, the trie descent makes a constant number of hops $(= log_B H)$, where $H$ is the size of the space of IDs, $B$ is the radix of the trie.

#figure(
  image(
    "../../src/tapestry_diagram.png",
    width: 50%
  ),
  caption: [Example connections of a node with ID 013. $L_i$ is the a connection at level $i$. (Note all connections are not shown for brevity)]
)

== *Node Deletion:*
- For graceful deletion, a node first identifies the closest node to its own ID, which then serves as its replacement during the deletion process.
- The replacement node updates routing tables and fills any gaps left by the departing node, ensuring continuity and network integrity during deletion.
- Node deletion involves three key steps: *RTUpdate*, to update routing tables; *BPUpdate*, to update backpointers; and *BPRemove*, to remove the departing node from others' backpointers.
- Random assignment of IDs gives an expected $O(log n)$ nodes to be updated after deletion (refer #link(<appendix>, text("Appendix"
, fill: blue))).
== *Add Object:*
- Objects can be inserted into the network from anywhere (like a *distributed hash table*)
- Ensures redundancy by invoking the `StoreObject()` RPC on upto two other nodes (giving a redundancy factor of 3), selected by scanning the routing table. These nodes then replicate the object in their respective local maps.

== *Object Publish:*
- The `Publish()` function is periodically invoked every 5 seconds in a separate `go` routine.
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

#pagebreak()
= Testing
== Stress Testing
The methods Route(), Insert(), Delete(), Publish(), FindObject(), and Unpublish() were thoroughly validated through automated tests. These tests were committed to a separate git repository and can be reviewed by checking out the testing fork.

== Performance Scaling Results
Reponse times for `Route()` and `Insert()` calls were measured for various sizes of the network.

#figure(
  image(
    "../../route.jpg",
    width: 95%
  ),
  caption: [Reponse time for `Route()`]
)

#figure(
  image(
    "../../insert.jpg",
    width: 95%
  ),
  caption: [Reponse time for `Insert()`]
)

The plots clearly indicate that `Route()` operates in constant time, whereas `Insert()` exhibits a growth trend that falls between linear and quadratic on the log plot, giving a complexity of $O(log^2 n)$
#pagebreak()
#h.title("Appendix") <appendix>
_Claim: The expected length of the longest common prefix with $n$ random strings is $O(log_(|Sigma|) n)$ where $Sigma$ is the alphabet_

_Proof:_
To get an upper bound on the length of the longest common prefix, it is convenient to assume the strings have inifinite length.

Consider just one string, the longest common prefix (_lcp_) with another random string follows a geometric distribution.

$ Pr{"lcp" >= m} = 1/(|Sigma|^m) $
As atleast the first $m$ characters must match.

This is a geometric distribution. Thus the final quantity is the max of $n$ i.i.d. geometric variables ($L_i$):

$ Pr{L_i < m} = 1 - 1/(|Sigma|^m) $

Thus, 
$ Pr{max(L_1, dots L_n) < m} = product Pr{L_i < m} $
$ = (1-1/(|Sigma|^m))^n $

$ => Pr{max(L_1, dots L_n) >= m} = 1 - (1-1/(|Sigma|^m))^n = f(m) $

The required probability is:
$ sum_(i=1)^(infinity) f(i) $
$ approx integral_(1)^(infinity) f(x)d x $

By approximating  $ 1 - 1/(|Sigma|^x) approx 1-e^(-x(1-1/(|Sigma|))) $
And using the standard integral $ integral_0^1 (1-(1-u)^n)/u d u = 1/1+1/2+dots 1/n approx ln n $
Gives us $E["lcp"] approx log_(|Sigma|)(n)$