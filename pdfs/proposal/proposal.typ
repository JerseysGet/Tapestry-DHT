#import "../header.typ" as h

#show: h.my-template
#h.title("Decentralised overlay network - Tapestry")
#h.subtitle("Project Proposal - Team 26")
#h.author("Rohan Sridhar (2022101042) Mohammed Faisal (2022101101) Shreyansh (2022111002)")

= Project Title
Implementation of a Decentralized Overlay Network Inspired by Tapestry

= Problem Statement
In large-scale distributed systems, efficient routing and resource location are critical challenges. Traditional fully connected networks are impractical due to high storage and maintenance costs, while unstructured peer-to-peer (P2P) networks suffer from inefficient search mechanisms. Tapestry, a structured overlay network, addresses these issues by providing scalable, fault-tolerant, and efficient routing with prefix-based matching.

This project aims to implement a decentralized overlay network inspired by Tapestry. The system will support efficient message routing, dynamic node membership, and resilience against node failures while ensuring fast lookups in a scalable network.

= Framework and Technologies
- *Programming Language:* Go (Golang)
- *Communication Protocol:* gRPC 
- *Data Structures:* Prefix-based routing tables
- *Hashing Mechanism:* SHA-1 or similar
- *Storage (Resources):* In-memory or simple file-based storage for node states

= Project Objectives
== *Implement Node Discovery & Routing:*
   - Nodes should be able to join and leave dynamically without breaking the system.
   - Efficient routing using prefix-based forwarding. $\O(log n)$ hops for a routing request.

== *Resource Location & Lookup:*
   - Implement mechanisms for storing and locating resources.
   - Ensure lookups occur in logarithmic time complexity.

== *Fault Tolerance & Adaptability:*
   - Handle node failures by reconfiguring routing tables.
   - Implement redundancy mechanisms to ensure continued operation despite node failures.

= Deliverables
- Implementation of a Tapestry-inspired overlay network with routing, resource lookup, and fault tolerance.
- A written report detailing the implementation approach, technical challenges, and results.
- A presentation and demonstration showcasing the working system and its capabilities.

