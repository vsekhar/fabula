# Motivation

The web started out as a decentralized collection of academic institution servers making documents available to each other. It has [shifted to a centralized model](https://hackernoon.com/the-evolution-of-the-internet-from-decentralized-to-centralized-3e2fa65898f5) consisting of large information services managing silos of data. This has some [drawbacks](https://bdtechtalks.com/2017/10/27/why-does-the-centralized-internet-suck/).

The [Distribute Web](https://hacks.mozilla.org/category/dweb/) (Dweb) movement arose to counter this centralization. Dweb projects like [IPFS](https://ipfs.io/) replace many of the web's naming and routing protocols with global content-addressable storage distributed among a set of decentralized servers. This comes at [a cost in bandwidth and availability](https://hackernoon.com/ipfs-a-complete-analysis-of-the-distributed-web-6465ff029b9b).

This raises the question: are only two architectures possible? Is it a choice between the efficient but centralized web of today and the fully decentralized but inefficient approach proposed by Dweb?

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