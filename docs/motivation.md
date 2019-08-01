# Motivation

The web started out as a decentralized collection of academic institution servers making documents available to each other. It has [shifted to a centralized model](https://hackernoon.com/the-evolution-of-the-internet-from-decentralized-to-centralized-3e2fa65898f5) consisting of large information services managing silos of data. This has some [drawbacks](https://bdtechtalks.com/2017/10/27/why-does-the-centralized-internet-suck/).

The [Distribute Web](https://hacks.mozilla.org/category/dweb/) (Dweb) movement arose to address this centralization. Dweb projects like [IPFS](https://ipfs.io/) replace many of the web's naming and routing protocols with global content-addressable storage distributed among a set of decentralized servers. This comes at [a cost in bandwidth and availability](https://hackernoon.com/ipfs-a-complete-analysis-of-the-distributed-web-6465ff029b9b).

This begs the question: are only two architectures possible? Is it a choice between the efficient but centralized web of today and the fully decentralized but inefficient approach explored by Dweb?

## What is the web?

To address this, we need to break down the web, not into its constituent technologies but into the jobs being performed. Specifically, the web is a stack of technologies and standards

 * Naming data (URLs)
 * Efficiently discovering routes to data (DNS, BGP)
 * Reading data (GET, HEAD)
 * Efficiently and securely transferring data (HTTPS)
 * Representing data (HTML, CSS, Javascript)
 * Writing and modifying data (POST, PUT, DELETE, PATCH)

Standardization enables clients and servers who are not previously acquainted or integrated with each other to nevertheless share and operate on data.

These features have unleashed enormous growth in information systems that would not have been possible in a world consisting of proprietary systems and protocols.

However there is a gap.

## What is missing?

The standardization of the web benefits only communication between one client and one server.

An application requiring 

