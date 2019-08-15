# Design

This document captures open and resolved design decisions.

## Resolved: Shared or separate timestamp requests (shared)

Does each server commiting a transaction obtain its own TrueTime timestamp from infrastructure, or is there one party that obtains a shared timestamp? --> shared.

Timestamps are signed by infrastructure so can't be forged. A _happens after_ relation can be obtained by including a token \[chain\] in the timestamp request (at the cost of a roundtrip). Each server can thus validate that the timestamp came from infrastructure. Each server also validates that it is committing a transaction with a timestamp greater than all previous timestamps for those data elements.

## Open: preventing backdating attack

Can a server maliciously backdate a transaction? Not unlaterally. No one will commit a transaction before the timestamp of any existing data record.

What about disjoint sets of servers? Can an attacker use a few servers that are "behind" to backdate a transaction relative to a few servers that are "ahead"? Yes.

Proposal: servers in phase 1 can provide a salt to the client, client provides this salt to timeservers, which sign it along with the timestamp. Timestamp is sent back to servers.

## Open: Servers that don't know about `Consistent-*`

Servers that haven't been upgrade will process PUT/PATCH/etc. commands immediately. This may be contrary to the semantics the client wants. Need a way to cause such servers to error out. A single unknown header may not do it. Do we need an `OPTIONS` request for every server to be safe?

New verb? MPUT and friends?

## Idea: Consensus service

[Chubby paper](https://static.googleusercontent.com/media/research.google.com/en//archive/chubby-osdi06.pdf) envisioned and rejected a more general _consensus_ service, rather than the lock service they implemented.

> In a loose sense, one can view the lock service as a way of providing a generic electorate that allows a client system to make decisions correctly when less than a majority of its own members are up. One might imagine solving this last problem in a different way: by providing a “consensus service”, using a number of servers to provide the “acceptors” in the Paxos protocol. Like a lock service, a consensus service would allow clients to make progress safely even with only one active client process; a similar technique has been used to reduce the number of state machines needed for Byzantine fault tolerance [24]. However, assuming a consensus service is not used exclusively to provide locks (which reduces it to a lock service), this approach solves none of the other problems described above.

Possible "layers" of a service:

 * TrueTime ranges: [tt.earliest, tt.latest]
 * TrueTime timestamps: tt.timestamp
   * Reduce TrueTime range to a single timestamp by performing a commit wait
 * Consensus service:
   * Establish and record encrypted payloads comprising transactions
   * Trusted third-party oracle for commitment and ordering
   * Provides recovery path without requiring servers to properly behave
   * Servers only need to properly prevent conflicts within _their_ locally stored data

### Flow

> TODO: Add representation of cross-server dependencies

 1. Client generates a `Consistent-Id`
 2. Client issues commands to servers with `Consistent-Id` header, and sends transaction initiation to consensus service with same header
 3. Servers reply to client and forward their own `Consistent-Token`s to consensus service
 4. Consensus service appends `Consistent-Token`s to commit record keyed on `Consistent-Id`
 5. Client requests commitment of transaction with all servers
 6. Servers acquire locks and submit PREPARED message to consensus service
 7. Consensus service appends PREPARED messages to commitment log
 8. Consensus service performs commit wait
 9. If majority of servers reply 


Open questions:

 * Payloads available to everyone?
 * Payloads queryable? For how long?
 * Scale:
   * Consensus service requires storage (logging) that scales with number of transactions (for however long records need to be kept)
   * Lock service requires storage that scales with number of locks
