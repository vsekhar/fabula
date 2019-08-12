# COMMIT

COMMIT is a proposal to add global consistency and atomic operations to the web platform.

## Summary

Clients can read and provisionally write data to servers using existing web primitives (`GET`, `POST`, `PATCH`, etc.). Multiple requests are tied together using a new `Consistent-Id` header which supplies a common client-generated value to all servers across  requests.

To atomically commit changes across multiple servers, the client initiates a round of Paxos among the servers with a new web verb `COMMIT`. Paxos is a fault-tolerant consensus algorithm that brings a set of participants to agreement on a particular monotonically increasing value. In this case, the value being agreed upon is a global timestamp representing the moment at which the transaction is applied. The timestamp is made externally consistent via a new public TrueTime service.

A series of changes across overlapping and disjoint sets of servers can in this way be made globally consistent without onerous coordination or the centralization of datastores.

This proposal focuses exclusively on achieving global consistency and atomicity. Other web technologies are assumed for establishing identity (certificates and signing authorities), security (TLS), and data models (WebDAV, JSON). Application-specific semantics (including applicable standards) are also likely necessary.

## Read next

* [Motivation](motivation.md): more details on where `COMMIT` fits within the set of distributed web technologies
* [Protocol walkthrough](protocol.md): a walkthrough of a simple transactional round using `COMMIT`

