## Pending

* Each batch tree is a sparse Merkle Tree
  * Capable of producing proof of inclusiong and _proof of non-exclusion_
  * This is needed to support cancellation (below)
  * Clients submit a doc, server returns notarization = hash(doc, server_salt), server_salt
  * Client can cancel notarization by submitting doc, server_salt
    * Server backs out notarization = hash(doc, server_salt)
    * Server verifies that notarization exists (checks last XXs worth of cache)
    * Server fails if notarization not found
    * Server logs
  * Clients cancel notarization
  * Clients cancel the transaction by logging the notarization and salt to produce a cancellation
  *

* Building a sparse merkle tree at scale
  * Have many candidate subtrees in flight
  * Servers submit entry to batchers at a given prefix
    * Longer prefixes when busy, shorter prefixes when not
  * Batchers gather prefixes and write a subtree that "claims" the range denoted
    by the prefix
  * Conflict: batcher receives two requests containing subtrees for prefix 'aa'
    * Each request "claims" the open interval [aa*,ab*)
    * Drop the smallest one (the one with the fewest entries)
      * The dropped requester tries again later with same batcher OR drops one
        prefix down
    * Write subtree with prefix to pubsub, submit upwards
    * Lazily write to GCS
  * Top level batch is submitted for sequencing
  * Sequencing uses atomic write of GCS
  * Startup:
    * All requests go to "" (empty) prefix, claiming the whole tree
    * First request succeeds, other requests conflict and fail
    * Other requests drop down a prefix to [0-f] and try again

* Supporting cancellation
  * Inspired by Revocation Transparency: https://github.com/google/trillian/blob/master/docs/papers/RevocationTransparency.pdf
    * Maintain a separate Sparse Merkle Tree to track revocations
    * Log changes to SMT to a CT-style log
  * So Fabula can maintain a time-bound or sharded SMT
  * Could also use a bloom filter
    * No false negatives but some false positives
    * Matches requirements for transaction cancellation (correctness, but with
      some spurious aborts)
    * Bloom filters can be merged monotonically, so they can be updated in
      parallel and conflicts reconciled without data loss (bitwise OR)
  * Transaction starts with a logged record and a duration
    * Hash = document + salt
    * Expiry = timestamp + duration
  * Participant can submit document and salt to "revoke" the entry, adding hash
    to the bloom filter
    * Semantics of "revoke" is application-dependent
    * Bloom filter operator (could be separate from log operator) returns with a
      log hash immortalizing the state of the bloom filter in which record was
      revoked (client can check)
  * Bloom filter operator must merge updates if conflicts occur
  * Bloom filter could be time bound based on time of entry being revoked or
    time of expiry.
  * Bloom filter needs to be periodically rotated, at which point the previous
    one is frozen.
    * Bloom filter can be merged into the log: once frozen, we can test the hash
      of each entry in the log against the frozen Bloom filter that covers its
      timestamp
      * BUT: does this mean every expirable entry has to expire with its
        Bloom filter?
    * If it is present in the Bloom filter, we can mark that entry revoked and
      link to the log entry that captures the Bloom filter when it was revoked
      * False positives might not matter to some applications that just want the
        log entry.
    * If it is not present in the Bloom filter and the entry can be marked as
      valid
    * Entries with active Bloom filters are fully logged and provable, but are
      in an unknown revocation state
      * This might not matter for some applications, they just want the log.
  * Bloom filters are built by batchers, a bloom filter is written for each
    sequence entry corresponding to all entries in the batch tree below.
  * Could also use a Sparse Merkle Tree (SMT)
    * Need a write-optimized SMT

* Combine contexts in batcher and sequencer
  * Keep trying until all contexts timeout, or some server-set maximum timeout
    * Return grpc:4 deadline_exceeded or http:504 gateway timeout
    * Context cancellation is a polite request, not a semantic requirement
  * Batcher and sequencer detect quickly when they are too busy, don't try:
    * Return grpc:8 resource_exhausted or http:429 too many requests
    * Backoff is to a longer prefix

* Group by source IP:
  * Alternate nibbles from IP and notarization hash
    * IP: abcdefgh (v4) or abcdefghijklmnopqrstuvwxyzabcdef (v6)
    * Hash: 123456789...
    * Batch string: a1b2c3d4e5g6h7...
    * Route upward batch requests to shrinking prefix of batch string
  * This causes batches to aggregate requests from a given network or
    subnetwork, i.e. parts of the internet.
    * So intermediate caches along the way can grab responses and hold them for
      downstream clients that are on their particular slice of the internet.
    * Increases probability that clients on your network will access similar
    batches, motivating the deployment of caches
  * Each 4-bit nibble (one hex character) gives 16 values.
  * Worst case, ISP with many clients uses a single NAT'd IP:
    * IP: ########
    * Batch string: #1#2#3#4#5#6#7...
    * First nibble is same for all ISP clients, doesn't disperse any traffic
    * Second nibble can disperse traffic to 16 hosts
    * Third nibble does nothing
    * Fourth nibble disperses to 16 hosts (256 total)
      * If written at this level, ISP need only cache 256 batches
      * Is this better?
  * BUT: server handling '#' will get bombarded, unless batchers remember a
    level, but then they will end up writing at the lowest level which will
    almost certainly lead to too many batches overall
  * Issue is IP space is VERY spiky in terms of traffic (due to NAT, unused
    ranges, etc.)
    * Need different way of mixing
    * For each nibble, hash a nibble of IP and a nibble of notarization hash
    * Batch string: hash(a1), hash(b2), etc...
    * For a high-traffic fixed IP, each hash will contain 16 possible values
      (same as if nibble of notarization hash was used directly)
    * Doesn't do anything for grouping...
  * Reality is including IP, a non-uniformly distributed space, in the grouping
    logic at all will necessarily lead to hotspots dependent on the intensity
    of grouping in the IP space (e.g. NAT).
  * Maybe nix in favor of custom proxies based on the entries and sequences you
    want to track and evergreen.
    * Mechanism to pull proofs "through" a caching proxy
    * Not just caching, if data expires, need a proxy that will store data for
      all entries and sequences submitted to it
    * Verify by running verification against a proxy data store rather than the
      global one
        * Proxies can have history the global store doesn't, and so can prove
          entries older than can be proved by the global store
  * Privacy:
    * Source IPs are still never stored, they are only used at the time
      an entry is submitted to route upward requests among batchers
    * The resulting batches are thus arbitrarily grouped with other requests
      from similar IP prefixes
    * However, reading batches tells you nothing about the IP prefix used to
      create it (though its height in the batch tree might)
    * You must already know the documents being notarized to learn anything from
      the group of notarizations in your pack
      * If you do know other documents, you could learn that that they were
        notarized from an IP with a common prefix to your own
      * How many common bits are revealed depends on the depth of the batch in
        the batch tree

* Peer forwarder:
  * Listens on a port for introductions via Serf
  * Maintains member list (in Serf)
  * Builds hash ring
  * ForwardRequest(key string, req interface{}) (resp interface{}, error)
  * Broadcast(msg interface{})
  * RegisterRecentBroadcastCallback(func(msg interface{}))
  * Leaderish (nope, see Sequencer):
    * Peer can ask peer forwarder for its fallback(s) (i.e. who would take over
      if that peer died), peer forwarder determines this via hashring
    * Peer can forward state to fallback(s) via user events
    * Peers register fallback handler (possibly many) and maintain state cache
    * State cache is used when creating new sequence writer
    * Forget it. Just have whoever writes a new sequence batch broadcast the
      sequence high-water mark to everyone (Sequencer below)

* Sequencer:
  * Every node has one, uses it when a request with prefix=* arrives
  * Batches requests, but with only one concurrent handler
    * Need backpressure to push lower nodes to attempt at longer prefixes
    * Max number of entries per batch? buffered channel?
  * After writing each new sequence block, broadcast the highwater mark to
    sequencer on all other nodes
  * Sequencer listens for broadcasts and updates its HWM if it gets a higher one
    (it means some other sequencer has successfully written a block)
  * This way any node is ready to take up being the sequencer
    * After sequencer failure (and potentially the failure of other nodes in the
      zone), a handful of nodes might try to become sequencer before memberlist
      settles
    * While they will contend at the storage layer, they will each at least be
      able to attempt with up-to-date information.
    * Sequencers can perhaps retry 2-3 times if they get storage conflicts so
      that they don't have to error out a bunch of requests down the tree.
  * HWM includes seq_no, hash, timestamp of batch written, everything the next
    sequencer needs to write the next batch

* Client-side proving:
  * Each notarization request returns a sequence entry and a path through the
    tree of batches
  * Fetch all batches to verify inclusion in batch tree
  * Compute peaks at entry.seq_no and fetch those sequence entries
  * Compute indices of consistency proof to some published recent.seq_no and
    fetch those sequence entries
    * might have some throughput issues if lots of people want to prove using a
      recently added sequence entry.
    * sequence entries are pushed to all servers, they could serve recent ones
      locally

* Proofs are sequences of commands to a hashing stack machine
  * hash b, hash b, hash b, push, hash b, pop, hash arg0, push, ...
  * push and pop commands navigate up and down tree
  * arguments are drawn from a Sequencing consisting of a path down the batch
    tree
  * See also: [authentication octopus algorithm](https://eprint.iacr.org/2017/933.pdf)
    * Linked from [Quadrable README#strands](https://github.com/hoytech/quadrable#strands)

* Proofs can be evergreened by requesting consistency-proof after consistency-
  proof
  * Service can guarantee that individual entries live for at least 1, 5, 10, 30
    years
  * Clients can log, later obtain entry proof up to some current seq_no_1
  * Clients can later obtain consistency proof from seq_no_1 to a future
    seq_no_2
  * Clients can keep evergreening their proof by getting and storing more
    proofs
  * Clients take on storage costs proportional to how long they want their
    proofs to live, relatively small O(MB), distributed among many entities
  * Clients can batch up proofs: prove N entries up to seq_no_1, then just keep
    evergreening the sequence number with additional consistency proofs
  * In this way the client "mirrors" a sub set of the whole tree relevant to
    their entries.
  * Chain of consistency proofs is a kind of skiplist
  * With client-side proving, lots of people can mirror the log and provide
    proof evergreening
  * Log stores a starting point {seq_no, hash} of the last entry that expired
    * Used to start verification of the remaining entries
    * Every month that starting point is advanced according to the expiration
      time
    * Entries before the starting point can be deleted
    * Or just delete everything prior to a given sequencer entry, retain one
      sequencer entry just prior to the current expiration time (can be verified
      against announced policy).

* Services can be provided to maintain evergreening for clients
  * Notarize "through" the service
  * Service tracks all notarizations and pulls proofs and sequences to prove
    only those entries
  * Service can charge for this
  * Essentially charging for continuous (as long as you pay) access to proofs
    for your specific subset of entries from the global log
  * Still universally causally related.

## Done

* GCE and MIGs instead of Kubernetes
  * Cheaper:
    * 3 x f1-micro + GKE = USD $88.33/month
    * 3 x f1-micro + MIG = $27/month
    * Source: https://cloud.google.com/products/calculator#id=62d37a1b-c646-4c0a-83f0-64a3d51d62b5
    * Load balancers are billed in either case: https://cloud.google.com/kubernetes-engine/docs/tutorials/http-balancer#background

* Batching:
  * all handlers in a task do blocking sends of each entry to one global
    unbuffered channel, with a timeout
    * handler returns Resource Exhausted (grpc: 8, http: 429)
    * upstream task increases prefix length and tries again (to another server)
    * TODO: need to encourage coalescing, write to one of an array of channels
      with a preference to earlier ones, or vary the number of batchers reading
      off the one big channel by reporting back pressure to the batch pipeline
  * N batcher goroutines are started, do non-blocking reads off the channel
    until nothing left to be read or max batch size reached
    * ensures entries don't wait and also provides some batching
    * if batcher queue empty and non-blocking read is empty, then do blocking
      read to wait for a value (or die to scale down the number of batchers)
  * batcher writes batch, then acks all the requests
  * batcher goes back to do more non-blocking reads and build another batch
  * handler

## Rejected

* Hierarchy of logs:
  * XXX: Rejected in favor of "proof evergreening" where clients build custom
    subordinated logs for the entries they want to ensure they can prove.
  * entries first go to a 1-day log and are expired quickly
  * clients must evergreen from 1-day to N-day, and eventually to 10-year log
  * clients cannot log directly to 10 year log, can only log a digest from a
    lower level (e.g. 5 year log) into the 10 year log
  * upgrades allow infrastructure to batch, 10 year log doesn't grow quickly
  * clients can do the work up front for the duration of relevance of the data
    they logged.
  * BUT: what is being logged in the higher level logs that allows them to
    serve older requests? Can the 10-year log prove all requests in the last
    10 years?

* Access control --> "work token"
  * XXX: Rejected in favor of expiring global log and paid evergreening sub logs
  * Use something like HashCash: client must include with request a counter and
    timestamp that satisfy:
    * timestamp within +/-5s of server local clock (not notarization timestamp)
    * hash(document, hashcash_timestamp, base64(counter)) has 20 leading bits
      that are zero (1 in 2^20 chance, or one in 1 million)
      * could break down to sub-problems to reduce variability
    * requests that fail HashCash check are returned with an error specifying
      the current server timestamp and the number of leading zero bits required
      to notarize something
      * failures do not store anything or propagate to other servers or to the
        sequencer
  * Issue is double-spending
    * DDoS:
      * some central place can generate stream of tokens for a fixed document
        for timestamps in the near-future
      * stream them to very many clients (e.g. botnet)
      * have them all log the fixed document at the appropriate timestamps
      * trying to track tokens only exacerbates hotspot in the event of large
        scale token re-use.
    * make tokens depend on client identity? can't use IP because client may
      not know it behind a NAT
    * client needs to get a nonce from server, or even just its own IP
      * server needs to track nonces and reconcile...
    * client needs to generate a private key, use it to sign its request and
      include the public key
      * server verifies signature
      * server verifies that public key hasn't been recently used a whole bunch
      * servers can gossip popular public keys and shut them down
      * this prevents a central server from signing a payload once and
        distributing it
      * this doesn't prevent a bot-net from generating new keys per request, but
        generating these keys and signing the payload is itself the proof of
        work.
      * could we do this with just checking that the document isn't too common?
        * but then we'd have to store the document
      * this check isn't that helpful since by the time we've noticed a key/doc
        is super common, it's already been logged a bunch of times and the
        central attacker can move on to the next document
