# Design thoughts

Centralized databases require you to trust one specific entity. High throughput, high risk.

Distributed databases (like bitcoin) allow you to trust no one. Low throughput, low risk.

Decentralized infrastructure (like PKI, certificate authorities, DNS) require you to trust *some* entities. High throughput, mostly low risk.

Verified decentralized infrastructure allows you to trust no one. High throughput, low risk.

## Notaries

Notaries are like certificate authorities: a trusted-ish group of authorities. Like with certificate authorities (in a post-[Certificate Transparency](https://www.certificate-transparency.org/) world), the work of notaries is logged, transparent and verifiable.

Participants can "bring their own notaries." This should not affect risk or correctness for other participants.

### Access control: Blinded tokens

Notaries sell tokens which are redeemed when a log entry is submitted.

Notaries can sell blinded tokens if they want to maintain their customers' privacy. See: [PrivacyPass protocol](https://github.com/privacypass/challenge-bypass-extension/blob/master/docs/PROTOCOL.md).

### Ordered logging: Merkle weaves

Notaries log to local verifiable write-optimized Merkle weaves.

### Timing: Notary clocks

Notaries utilize their own local clocks when notarizing.

Notaries can estimate `local_uncertainty` in their own clocks via some internal method.

A notary has no incentive to over-estimate `local_uncertainty` since it would only slow down its notarizations (see below).

### Uncertainty: Cross notarization

Notaries periodically log to other notaries in a trusted set every O(seconds). E.g. GCP, AWS and Azure notaries could log to each other.

Notaries generate a local log entry, then log that log entry with a remote notary, and log the remote notary's response locally again, etc. Notaries do this at random intervals and for random-length chains.

Notaries keep records of these cross-notarizations.

Notaries use the timestamp differences in cross-notarizations to estimate `global_uncertainty`.

Overall `uncertainty = sqrt(local_uncertainty^2 + global_uncertainty^2)`.

### Uncertainty: Verifying

Anyone can query for the latest cross-notarizations from a given notary and verify that notary's local estimate of uncertainty.

Each cross-notarization is logged at the remote notary, so they cannot be forged (without colluding with the other notary), and logged locally so they cannot be re-ordered.

### Notarization

Customers can ask a notary to notarize a 64-byte document, usually a hash of some other document.

The notary generates a 64-byte salt, a hash of the document+salt, a timestamp and a signature.

The notary performs a commit-wait to wait out its local estimate of `uncertainty`.

The notary returns the generated values to the customer.

### Summaries

Notaries periodically publish summaries of their log. This takes the form of a set of summaries of each of the underlying sharded Merkle trees. Customers can request the latest published summary, often a second or two stale.

Customers can also request a summary as of a specific timestamp. For this request, the log produces a summary where the last timestamp in each sharded Merkle tree is at or after the requested timestamp. This is used to prove to the client that the log will accept no log entries prior to the given timestamp, preventing the possibility a colluding log could "front run" the client's log entry. Requests for summaries at very recent timestamps are expensive and may incur delay.

### Proving summary inclusion

Customers can request that a proof be furnished from any Merkle tree summary to any future Merkle tree summary. The response contains the hashes needed to construct the later summary from the earlier one following the MMR scheme. This proves that the log has not discarded any entries summarized by the earlier summary.

Customers can also batch request a proof from a set of Merkle tree shard summaries to a future set of summaries, including for the entire sharded Merkle tree. The response consists of a set of proofs of each individual Merkle tree as described above. This proves that the log overall has not discarded any Merkle tree shards, and by extension has not discarded any entries in those shards summarized by the earlier summaries.

### Proving entry inclusion

Customers can request proofs of any log entry up to a specified summary of that Merkle tree shard. This proves the log has not discarded the entry.

### Multi-notary logging

A two step process can be followed by clients attempting to simultaneously log to multiple notaries.

First, the client obtains time hints from each of the notaries (or uses cached time hints). This allows the client to estimate its clock skew and latency to each notary.

Second, the client sends a logging request to each of the notaries with a timestamp slightly in the future for all of them.

Each notary verifies that the time stamp is near enough to its local time (not behind any existing log entries, not too far in the future to impact performance). Each notary logs the entry as of the specified timestamp. Finally, each notary commit-waits until it is sure it's own `Now() - uncertainty` is after the specified time before returning the notarization to the client. Slow notaries will have to commit-wait longer. Fast notaries possibly not at all.

The client has to keep trying, adjusting time stamps, until it is able to log an entry with the same timestamp at all notaries. Significant network jitter between client and notaries could affect this process, particularly if the client is attempting to use notaries in different regions.

> **Tickets**: an alternative is to have clients send the data and request a ticket with a timestamp from a notary. A ticket is not a notarization. The timestamp of the ticket is the earliest possible time the notary will subsequently accept a logging request. This gives clients a wider window in which to choose timestamps but it's unclear how the notary will maintain pending tickets in a verifiable way.

## Performance

Sharding to 256 trees produces ~2KB summaries. How fast are writes? ~400ms latency... (see `deploy/storage/perfdiag.txt`) 2 writes per second per shard? 512 QPS?

Sequential writes, with probing (80 threads) --> 645 ms/write:

```bash
$ go test -benchmem -run=^$ github.com/vsekhar/fabula/pkg/benchgcs -bench "^(BenchmarkAppendParallel)$" -benchtime 60s
goos: linux
goarch: amd64
pkg: github.com/vsekhar/fabula/pkg/benchgcs
BenchmarkAppendParallel-8            100         645055552 ns/op                 1.00 collisions/op             19.0 probes/op  108473856 B/op     10385 allocs/op
PASS
ok      github.com/vsekhar/fabula/pkg/benchgcs  65.935s
```

Databases may have same issue.

## Time ranges

Notarizations don't get timestamps, they get time ranges. Notaries wait until local time minus global uncertainty is past the second timestamp of the entry. Can drop entries into storage without sequencing. Looser notion of order, only among entries those ranges don't overlap.

How is it made verifiable? Lexicographic order? But how hashed? Delay until processed. Processing is serialized. Can we batch it? Do the intervals help us?

Can't key off intervals directly. Produces hotspot near most recent intervals.

## Time ranges and PubSub

* PubSub topic
* One subscription for each level (using filtering)
  * Level stored as an attribute of messages
* Packers collect messages at a specific level for some amount of time until a threshold is reached (time or number of messages)
* Packers order the messages they received by beginning and end timestamps
* Packers hash the summary of these messages
* Packers write the summary as an entry
* Packers append summary hash to metadata of included messages (using metadataGeneration to ensure consistency)
* Packers submit the entry to PubSub with attribute for next level up
* Packers acknowledge all included entries
  * It's ok for a given entry to be included in more than one upper level summary
* PubSub ensures each entry is included in at least one pack further up
* Top level subscription is essentially a stream of global summaries
* How high to go? How to know which one to read?
* Each entry contains in its metadata a pointer to a higher level entry that includes it

* Higher level entries (maybe the highest level one?) can be delievered strictly in order and has its previous
  * Should mid-level entries be sequenced?
  * Sequence based on level? E.g. each one has an ordering key that is shorter until the root one which has a single ordering key?
  * Ordering key is just a prefix, level of a message is just length of its prefix
    * Packers strip prefix down to increase the level of a message
    * How to pack across ordering keys?
    * Each pack might be ordering messages at multiple levels. OK?
  * Ordering guarantees that the arrival of a message with a given ordering key implies the processing of all messages before it
    * If a prior message is NACK'd, then the current message will be redelivered somewhere as part of the redelivery of that NACK'd message)

* Summary entries don't guarantee a timestamp HWM... There could be earlier entries still in the system.
  * Just look at the entry's attributes in GCS to see if it has been summarized
  * Follow entry attributes up as far as you like until you are satistifed of its inclusion
* What about pubsub timeout?

* Throughput
  * PubSub limits publish throughput to 1MB/s for each value of OrderingKey ([source - internal](https://docs.google.com/document/d/1DRIG8kkH8jqEBXRtJrp9vni0nE_uFtRFPPaeMuj1l6Y/edit#heading=h.3xuxqryq3rjf))
  * Message: 64 byte hash + 20 bytes timestamp + 16 byte message_id + 8 byte ordering key = 108
    * NB: PubSub _bills_ for min. 1000 bytes per message ([source](https://cloud.google.com/pubsub/pricing#message_delivery_pricing))
  * Goal of 1B events/s == 108GiB/s --> min. 108000 OrderingKeys --> 17 bit prefix at lowest level
    * Though likely rate limited by GCS

* Timestamps
  * Timestamps in pubsub are only available on the subscriber end, not the publisher end
  * Can't get add a timestamp to the message because it will not obey causality
    * PubSub ordering uses server-determined timestamp (`PublishTime`), not client timestamp (`SubmitTime`)

```bash
$ go run cmd/psops/psops.go -verbose -n 10 drain
...
2020/09/28 20:21:29 Drained Ft5gqaj6vSU... (ID: 1577689094528315, OrderingKey: Fg, SubmitTime: 2020-09-29 00:21:25.790383342 +0000 UTC)
2020/09/28 20:21:29 Drained Fhvr0pPQe8U... (ID: 1577689094528316, OrderingKey: Fg, SubmitTime: 2020-09-29 00:21:25.790640644 +0000 UTC)
2020/09/28 20:21:29 Drained FghMyfm3_BI... (ID: 1577689094528317, OrderingKey: Fg, SubmitTime: 2020-09-29 00:21:25.790326235 +0000 UTC)
```

* Put the entire tree in PubSub (put summaries in the data fields of PubSub)
  * Separate subscriber, filtering on summaries, that persist to GCS
  * Serve from GCS
* This exploits the semantics of PubSub message ordering and eliminates the need to check GCS for the last pack
  * CAVEAT: summaries lose information, once a lower level entry is acked, its contents are gone
  * CAVEAT: How do you get the last entry of a summary for chaining?
    * If a server goes down, need to find the last highest value
    * Can we enbed these in messages somehow? Server going down still results in discontinuity...
    * Need to persist this value
    * --> GCS...
    * Sequencing requires dependence on _ordering_ and _content_
    * PubSub supports no dependence between values
    * PubSub + ordering adds only ordering dependence, not content dependence
    * Submit a continuation entry? Can't. Continuation needs to get ahead of the entires it continues
    * Packers probe GCS to find last seq_no for their prefix, load it, cache its seq_no and hash
    * Packers write hex(prefix)-seq_no.pack with DoesNotExist condition, cache for next pack
    * Clients do binary search on a prefix to find an entry, knowing that timestamp ranges covered by each pack in a prefix sequence are distinct
      * Cache timestamp bounds along the way
      * Packs don't change, so can persist this index somewhere and load it to warm a client
        * hex(prefix)-timestamp.earliest-timestamp.latest.index
          * Contains records mapping {timestamp.earliest-timestamp.latest} --> seq_no
        * can be lazily written and can overlap (just used to build an immutable cache in client memory)
        * Indexer task?
    * PubSub handles aggregating messages and stably assigning them to packer tasks, ensuring packers are likely to be efficient in this way

------------------

Single-notary:

```http
Client:
  Document: <64 bytes>

Notary:
  Salt: <64 bytes>
  NotaryTimestamp: <Ts>
  UserTimestamp: <Ts>
  Notarization: <64 bytes> // SHA3512(Document, Salt, NotaryTimestamp, UserTimestamp)
  Signature: ...  // of Notarization
```

Establishes that according to Notary, Document existed as _as early as_ Timestamp. Since there is only one notary, Timestamp is an acceptable timestamp for Document.

Multiple notaries can be used in two phases. First the client has to guess a timestamp acceptable to all notaries. To aid in this, clients can ask notaries for timestamp hints:

```http
Client --> each notary:
  Document: null

Notary A --> Client:
  NotaryTimeHint: 15

Notary B --> Client:
  NotaryTimeHint: 22
```

Time hints are non-binding, not logged, and are intended only to help clients guess at timestamps. Clients can use time hints and round trip delays to guess a timestamp acceptable to all notaries. For example, if a client guessed a timestamp of 35, it will attempt to commit the document to all notaries via:

```http
Client --> each notary:
  Document: <64 bytes>
  DocumentTimestamp: 35
```

The notaries each verify locally that `Now < DocumentTimestamp`. The notaries then independently commit-wait until each can be sure ClientTimestamp is in the past per their own clocks. After the commit-wait, the notaries respond to the client:

```http
Notary A --> Client:
  Salt: <64 bytes>
  NotaryTimestamp: 21
  Notarization: <64 bytes> // SHA3512(Document, Salt, DocumentTimestamp)
  Signature: ...  // sign(NotaryTimestamp, Notarization)

Notary B --> Client:
  Salt: <64 bytes>
  NotaryTimestamp: 27
  Notarization: <64 bytes> // SHA3512(Document, Salt, DocumentTimestamp)
  Signature: ...  // sign(NotaryTimestamp, Notarization)
```

The document is now duly notarized with a timestamp of 35. Since each notary commit-waited until the notary could be (locally) sure no dependent document would receive a timestamp less than DocumentTimestamp, the timestamp of 35 is externally consistent.

Notaries can reject DocumentTimestamp that is in the past (violating causality) or too far in the future (requiring a lengthy commit-wait). Importantly, each notary can make this policy determination locally, according to their local clocks.

### Attacking causalilty

It is not possible for notaries to fake causal violations for other notaries.

It may be possible for notaries to detect that they are being audited and perform extra work to prevent causal violations (and otherwise permit or create causal violations for other clients). We would need a way to blind audits even if the source of the audit log requests (another notary) is known to the notary under audit.

## Protocol

### Multiple timestamping notaries

Each party has a set of notaries it trust. A document is not valid until it is timestamped and signed by a set of notaries that covers each of the participants.

Worst case scenario: each party trusts only its own notary.

Protocol:

* Client submits a document to be signed to each notary in parallel and obtains a _ticket_
* Notaries respond with a signed time _interval_: `{tt.Earliest, tt.Expiry}`
  * `tt.Earliest` is determined by the notary's local externally-consistent interval clock
  * `tt.Expiry` is determined by the notary's local policy, i.e. how long it will keep the ticket open waiting for the notarization to complete (e.g. 30 mins).
    * Note that this is not the same as `tt.Latest` since that is part of the external consistancy guarantees, which are not needed here.
    * `tt.Expiry` reflects only the longest the notary is willing to commit-wait up to some chosen future timestamp
* Client chooses `timestamp := max(tt.Earliest, ...)` among all notaries
* Client submits document and chosen timestamp back to notaries
* Notaries confirm document matches earlier request
* Notaries confirm chosen timestamp is after `tt.Earliest` from first request
* Notaries confirm chosen timestamp is before `tt.Expiry` from first request
* Notaries commit-wait to the chosen timestamp
  * I.e. each notary waits until it can locally assert `tt.Now().Earliest > timestamp`
  * Notaries that provided the earliest intervals will have to commit-wait the longest
* After commit-wait, each notary returns a signed document+timestamp
* Client submits document, with the chosen timestamp and signatures for that timestamp from all notaries, to other participants
* Participants can locally verify signatures of their chosen notary and thus trust the chosen timestamp is externally consistent.

In this way, each notary _locally_ enforces causality (won't accept a chosen timestamp earlier than that notary's local `tt.Earliest`).

#### Attacks

A malicious notary can:

* Provide a time interval far in the past to attempt to game causality
* Provide a time interval far wider than other notaries
* Provide a time interval far in the future than other notaries

In all cases, the other notaries will not accept a chosen timestamp earlier than _their own_ `tt.Earliest`. This guarantees their own notion of causality.

Malicious notaries can cause delays by refusing to accept a sufficiently-in-the-past timestamp. However this will be visible to the client. The client can see the malicious notary's unusually large `tt.Earliest` and decide to abort the transaction. Notaries can also set expiry times to ensure they are not hung up holding resources while a malicious notary dawdles. Clients can also see expiry times. If the time intervals `{tt.Earliest, Expiry}` provided by notaries result in a disjoint set, then the client can immediately determine that no timestamp is possible and can abort the transaction.

As a result, only _one_ honest notary will result in external consistency for _all_ participants. Dishonest notaries can delay or prevent notarization but will produce signed records of their attempts to do so.

## TODO

* Is a causal audit generalizable?
  * Can multiple notaries ganging together during normal operation also implicitly and continuously audit each other?
  * Maybe. See multi-notary protocol.
* How do notary expiry times interact with transaction expiry times?
  * Aren't they the same?
  * Can the client ask for a specific expiry time?
  * Then the client can produce from the notary an expiration certificate. I.e. _every_ notarization becomes an expirable element of a transaction.
  * If _any_ notary required for a transaction has issued an expiration certificate, then it will never be possible for that transaction to complete and so everyone can abort it.
  * Thus expirations have to be logged permanently...
  * Malicious notaries can abort away. But that is apparent. And is equivalent to a participant aborting away (since the participant demands the use of a notary).
  * Can a malicious notary fork history? Abort in one timeline, notarize in another?
* How do you compare notarizations from disjoint (sets of) notaries?
  * Client can identify a set of notaries that are behind and use them to violate causality
  * Notarize a doc with the ahead notaries, then notarize a dependent doc with the behind notaries
  * Causal audit?
  * Require pre-reasoning about relationships? I.e. require an overlapping notary?
  * Overlapping notary gains power...
  * --> Single notary... :\
  * **Make notarizations contain within them their timestamp so that dependent documents cannot be notarized at an earlier time**
    * But we are oblivious to document contents and format
    * Maybe the timestamp is special...
    * Turns into a vector clock... or a chain of documents
    * Every document has to be chained, cannot reason about order of events outside of chains :\
  * **Uncertainty is a global property**
    * Distributed federated protocol for notaries to check each other and come to agreement on a global estimate of uncertainty
    * Notaries are then (supposed to) use this global uncertainty when commit-waiting
    * Federation of notaries during a transaction still applies: even a single well-behaved notary in a transaction will ensure that transaction has a good timesetamp
      * Preserves local interest in and actionability on ensuring a trusted notary is used
    * So how do notaries agree on a global value for uncertainty?
* Esimating global uncertainty
  * Notaries can reliably check each other's clocks: Notary A has its own timestamp notarized by Notary B
    * Produces a signed statement of both clocks
  * Some distributed version of [Marzullo's algorithm](https://en.wikipedia.org/wiki/Marzullo%27s_algorithm)?
  * How do you secure a shared estimate of uncertainty?
  * Localize calculation of uncertainty
    * Notaries individually test each other, establish deltas and publish signed records of those bilateral (multilateral?) tests
    * Notaries review signed records of tests from other notaries and develop _their own_ estimate of _global_ uncertainty
    * Notaries have an incentive to ensure their own value for global uncertainty is accurate
    * NB: high uncertainty slows down the notary
  * Notary architecture - each regional instance consists of:
    * Time servers run chronyd, make use of Spanner or AWS Time Sync Service, try to establish as accurate a local sense of time and local uncertainty as possible
    * Log servers maintain log (primarily a database building a Merkle weave)
      * As an optimization, time servers and log servers can be combined if Spanner is used at the backing store of the log, however a non-Spanner implementation would be ideal
      * Non-Spanner log servers would need sharded append-only database (Vitess?)
    * Sync servers:
      * Learn of other notary instances
      * Periodically ping them for their sync log
      * Periodically log new sync records locally and with another notary (bilateral)
      * Periodically request proofs for sync records it obtains
      * Periodically update local estimate of uncertainty
    * Web servers respond requests to:
      * Notarize: append to log with timestamp, perform commit-wait
        * Incoming syncs from other notaries take the form of simple notarization requests, as a result, a notary is blind to sync requests as they are blind to all other notarization requests
      * Audit log: provide proofs from the log
      * Sync log: provide sync records from neighbor notaries
  * Notaries observe sync logs, start to establish "trust" in each other
    * Persistent outliers lose trust and become less influential
    * Sync algorithm
      * Each notary builds a graph of itself and other notaries
      * Each sync record (the notary's own or ones it gets from others) updates the edges in the graph
      * Graph edges accumulate uncertainty over time
      * Compute overall uncertainty
      * Is there a paper on this somewhere?
  * With the above, it doesn't matter if a new notary comes along (say on my machine), it won't have great reliability
  * Can a notary be surrounded malicious notaries and be steered away?
    * Notaries always use their local time, only their _uncertainty_ is affected by other notaries
    * So a notary surrounded by malicious notaries would just slow down, take longer to commit-wait before notarizing documents for its user
    * If the notary detects that _it_ is the persistent outlier among the cohort of other notaries, it may decide it has a bad clock and shut itself down
    * Notaries are _infrastructure_ and are thus _operated_ by someone
    * Being surrounded by malicious notaries makes the notary slow
    * So notary operators have an incentive to introduce their notary to other known trustworth notaries (like Google's or Amazon's) to keep their notary fast
    * Can malicious notaries "flood the zone" by their numbers?
      * Sure, but they'll all start off untrusted and so having little to no impact
      * How do they get trusted? By notarizing with timestamps in close proximity to other notaries, correctly logging those notarizations, correctly producing proofs of those log entries, etc...
      * By that point, your "attackers" have built a new and robust infrastructure for keeping time...
      * MFA: [https://xkcd.com/810/]
      * Remember: trust between notaries is different from trust between users and notaries
        * Notaries only trust each other to give them a good esimate of the time, to reduce their own local estimate of uncertainty
        * Users trust notaries to consistenly notarize, log and prove log entries

```http
Alice <--> George:
// Some arbitrary protocol that produces some consensus data that they wish to notarize
Data: Alice and George agree the bikeshed should be blue.

Alice --> AWS:
Document: <hash of Data and notaries>
Duration: 30s

AWS --> Alice:
Salt: <random bytes>
Notarization: <hash of Document, Salt>
Current-Time: 36
Expiry: 66

George --> GCP:
Document: <hash of Data and notaries>
Duration: 30s

GCP --> George:
Salt: <random bytes>
Notarization: <hash of Document, Salt>
Current-Time: 42
Expiry: 72

[George tells Alice he will consider the agreement valid if Alice can obtain a notarization from notary.gcp.com within the expiry time window]

George --> Alice:
Salt: <random bytes from GCP>
Notarization: <hash of Document, Salt>
Notary-Hostname: notary.gcp.com
Notary-Identity: <public key>
Notary-Timestamp: 42
Notary-Expiry: 72
Signature: <George signs Data, Notary-Hostname, Notary-Identity, Notary-Timestamp, Notary-Expiry>
// This signature is George's commitment that the transaction is valid if a signed notarization
// from notary.gcp.com can be produced with a timestamp before Notary-Expiry. When joined with
// such a signed notarization, no other statement from George is required to prove George's
// commitment.

// TODO: if Alice goes down, George may be in an undecided state.

Alice --> {AWS,GCP}:
Document: <hash of Data>
Salt: <random bytes from earlier>
Notarization: <hash of Document, Salt>
Timestamp: 61

{AWS,GCP} --> Alice: (after commit-wait)
Timestamp: 61
Signature: <Notary signs Notarization, Timestamp>

[Alice verifies her own notarization to AWS succeeded within the Expiry window]

Alice --> George
Salt: <random bytes from GCP>
Notarization: <hash of Document, Salt from GCP>
Timestamp: 61
Signature: <GCP signature of Notarization, Timestamp>
Signature: <Alice signs Notarization, Timestamp>
```
