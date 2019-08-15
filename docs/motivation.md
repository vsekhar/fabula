# Motivation

The web started out as a decentralized collection of academic institution servers making documents available to each other. It has [shifted to a centralized model](https://hackernoon.com/the-evolution-of-the-internet-from-decentralized-to-centralized-3e2fa65898f5) consisting of large information services managing silos of data. This has some [drawbacks](https://bdtechtalks.com/2017/10/27/why-does-the-centralized-internet-suck/).

The [Distribute Web](https://hacks.mozilla.org/category/dweb/) (Dweb) movement arose to counter this centralization. Dweb projects like [IPFS](https://ipfs.io/) replace many of the web's naming and routing protocols with global content-addressable storage distributed among a set of decentralized servers. This comes at [a cost in bandwidth and availability](https://hackernoon.com/ipfs-a-complete-analysis-of-the-distributed-web-6465ff029b9b). While batching and local signing permit the rapid creation of transactions, ensuring transactions are durably committed requires time for the transaction to "settle" into the distributed data structure.

This raises the question: are only two architectures possible? Is it a choice between the efficient but centralized web of today and the fully decentralized but inefficient approach proposed by Dweb?

## ACID properties and roles

The [ACID properties](https://en.wikipedia.org/wiki/ACID) are required for updates to data structures:

 * Atomicity: either all modifications to data are applied, or none are (even in the presence of failures, crashes, etc.)
 * Consistency: invariants are preserved at all points (even invariants that apply across multiple systems)
 * Isolation: transactions and their effects are well-ordered, and not interleaved
 * Durability: committed transactions remain commited (even in the presence of failures)

Distributed systems like block chains and hash graphs entrust the achievement of all four ACID properties to a single data structure. Since that data structure is distributed, _all servers_ become responsible for providing all four ACID properties to all clients.

A better breakdown is likely possible starting with the following observations:

 * Servers are interested in and will invest the bandwidth and cycles to manage _their own data_
 * Servers are interested in and will invest the bandwidth and cycles to server _their own users_

In a sense this is already a decentralized world. Why not entrust the attainment of ACID properties to the group of servers involved in each transaction?

A problem arises with a subtle consequence of one of the ACID properties: isolation.

It is natural to assume servers will want all transactions on their data to be locally consistent and locally well-ordered. Transaction `T1 = {clientA, serverX, serverY}` will need to be well-ordered among all transactions on `serverA` and `serverB`, and we can entrust these servers to coordinate amongst themselves to ensure that is the case.

However the effects of that transaction may then be seen by arbitrary and as-yet unknown other parties. For example, a subsequent transaction `T2 = {clientB, serverY, serverZ}` needs to be well ordered relative to all transactions on `serverY` including `T1` above, while at the same time being well-ordered relative to all transactions on `serverZ`.

As a result, the _implied cohort_ of all transaction is all past, current and future participants of any other transaction anywhere. That is, we require a robust _global_ ordering of transactions even though transactions are otherwise inherently local.

As a global property, ordering is in fact the _only_ one of the ACID properties that need specialized infrastructure to support. The other properties can be duly achieved via local capabilities on each server involved in each transaction. Trying to pile on all ACID properties into a single system like a block chain invites more headache than is warranted.

So let's build a global system for durably ordering ACID transactions, and nothing else.



TBD

 * A client is interested in seeing their transaction committed (otherwise why use the system at all?)
   * Responsibilities: driving a transaction forward, collecting replies/signatures, initiating commit
 * A server is interested in ensuring its data is safe and consistent
   * Responsibilities: verifying data integrity and consistency

TBD
 * One or more servers holding the data transacted upon
 * Trusted infrastructure that is blind to the contents of the transaction (similar to DNS)

We assign ACID properties as follows:

 * Atomicity: servers are responsible for atomic application of transactions
 * Consistency: servers are responsible for checking invariants
 * Isolation: servers are responsible for locking and preventing transaction interleaving, infrastructure is responsible for providing global externally consistent time stamps
 * Durability: infrastructure records opaque transaction IDs, timestamps, timeouts, and status


## What is the web?

To address this question, we need to break down the web, not into its constituent technologies but into the jobs being performed. Specifically, the web is a standardized stack of technologies for:

 * Naming data (URLs)
 * Efficiently discovering routes to data (DNS, BGP)
 * Reading data (`GET`, `HEAD`)
 * Efficiently and securely transferring data (HTTPS, GZIP)
 * Representing data (HTML)
 * Presenting data (CSS)
 * Interacting with users (Javascript)
 * Modifying data (`POST`, `PUT`, `DELETE`, `PATCH`)

By using standard technologies to perform these jobs, the web enables clients and servers who are not previously acquainted or integrated with each other to nevertheless share and operate on data. This capability has unleashed enormous growth in information systems that would not have been possible in a world consisting of proprietary systems and protocols.

However there is a gap.

## What is missing?

The standardization of the web benefits applications involving one client and one server (OCOS for short). This architecture is a natural outgrowth from the web's origins as a system for publishing and retrieving academic documents.

As we've added capabilities to the web, the OCOS assumption has remained largely unchanged. While many "servers" these days are in fact distributed systems of servers, they are almost universally part of the same organization, or "administrative domain." Building applications that seamlessly make use of multiple servers administered by multiple organizations currently requires extensive a priori and proprietary integration, preempting the benefits of a standardized and open web.

[Diagram showing OCOS single machine, OCOS "distributed" single machine, OCMS proprietary, OCMS open (COMMIT)]

As a result, the _business_ architecture of the web is largely today confined to data silos. Large distributed systems depend on a common administrative domain for trust, integration and consistency. This means the data used by any given application (an email service, mapping service, or social network) must be in the possession of a single organization. 

## COMMIT

COMMIT is a proposal to add a simple but powerful primitive to the web to unlock the development of genuinely multi-server (that is multi-administrative domain) applications that are secure, consistency and, most importantly, open.

Specifically, COMMIT adds a new HTTP verb, the Paxos 