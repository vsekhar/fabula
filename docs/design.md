# Design

This document captures open and resolved design decisions.

## Resolved: Shared or separate timestamp requests (shared)

Does each server commiting a transaction obtain its own TrueTime timestamp from infrastructure, or is there one party that obtains a shared timestamp? --> shared.

Timestamps are signed by infrastructure so can't be forged. A _happens after_ relation can be obtained by including a token \[chain\] in the timestamp request (at the cost of a roundtrip). Each server can thus validate that the timestamp came from infrastructure. Each server also validates that it is committing a transaction with a timestamp greater than all previous timestamps for those data elements.

## Open: preventing backdating attack

Can a server maliciously backdate a transaction? Not unlaterally. No one will commit a transaction before the timestamp of any existing data record.

What about disjoint sets of servers? Can an attacker use a few servers that are "behind" to backdate a transaction relative to a few servers that are "ahead"? Yes.

Proposal: servers in phase 1 can provide a salt to the client, client provides this salt to timeservers, which sign it along with the timestamp. Timestamp is sent back to servers.