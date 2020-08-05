# Fabula

> _Fabula (Russian: фабула, IPA: [ˈfabʊlə]) is the chronological order of the events contained in a story, the "raw material" of a story._ -[Wikipedia](https://en.wikipedia.org/wiki/Fabula_and_syuzhet)

Fabula is a proof-of-concept stack of services and protocols for verifiable decentralized action on the web.

It consists of the following components:

 * Notary: a trust-nothing verifiable notary and timestamping service
 * Button: a timed atomic decision primitive
 * Consensus: a consensus protocol for multiple parties to agree on some action

## Notary

A notarization consists of three stanzas:

 1. A `ticket` generated by infrastructure
 1. A `document` generated by the client
 1. A `certificate` generated by infrastructure

Infrastructure generates and sends to the client a `ticket` with the following logical structure:

```http
X-Notary-Ticket-Timestamp-Range: <UTC nanosecond timestamp earliest>, <UTC nanosecond timestamp latest>
X-Notary-Ticket-Predecessor: <64 byte hash of "previous" ticket(s)>, ...
X-Notary-Ticket-Hash: <64 byte hash covering Timestamp-Range and Ticket-Predecessor(s) in order>
```

The infrastructure logs `Ticket-Hash` and can prove any such hash on request.

The client generates a `document` and sends it to the server with the following logical structure:

```http
X-Notary-Document-Timestamp: <UTC nanosecond timestamp within range of Timestamp-Range>
X-Notary-Document-Data: <64 bytes base58 encoded>
```

The server returns a `certificate`:

```http
X-Notary-Document-Hash: <64 byte hash covering Ticket-Hash, Timestamp, Data, Salt>
X-Notary-Certificate-Predecessor: <64 byte hash of "previous" certificate(s)>, ...
X-Notary-Certificate-Hash: <64 byte hash covering Document-Hash, and Certificate-Predecessor(s) in order>
```

The infrastructure logs `Certificate-Hash` and can prove any such hash on request.

The client can locally verify `Hash-Certificate`. The client can persist and pass around the full notarization to prove the validity of some document. Readers can locally validate that `Document-Hash` covers the document in question, and that `Certificate-Hash` covers a validly-constructed timetsamp as well as `Document-Hash`. Readers can remotely validate `Certificate-Hash` in the notary log.

Tickets are stored and logged in a verifiable log published by the infrastructure. Documents and certificates are stored in a separate log. Every certificate maps to and consumes one document, and every document maps to and consumes one ticket. However, some tickets can go permanently unusued, and some documents can go permanently uncertified (if for example it never makes it back to the infrastructure to be logged and certified).

The full notarization can be internally validated using the hashes. In addition, `Hash-Ticket`, `Hash-Document` and `Hash-Certificate` can be externally validated in the appropriate logs.

```http
X-Notary-Ticket-Timestamp-Range: <UTC nanosecond timestamp earliest>, <UTC nanosecond timestamp latest>
X-Notary-Ticket-Predecessor: <64 byte hash of "previous" ticket(s)>, ...
X-Notary-Ticket-Hash: <64 byte hash covering Timestamp-Range and Ticket-Predecessor(s) in order>
X-Notary-Document-Salt: <64 bytes base58 encoded>
X-Notary-Document-Timestamp: <UTC nanosecond timestamp within range of Timestamp-Range>
X-Notary-Document-Data: <64 bytes base58 encoded>
X-Notary-Document-Hash: <64 byte hash covering Ticket-Hash, Timestamp, Data, Salt>
X-Notary-Certificate-Predecessor: <64 byte hash of "previous" certificate(s)>, ...
X-Notary-Certificate-Hash: <64 byte hash covering Document-Hash, and Certificate-Predecessor(s) in order>
```

The protocol has the following properties:

 * Infrastructure pre-commits to a timestamp range, reducing its ability to fudge timestamps
 * Infrastructure publishes timestamp range, and the range becomes part of the final notarization, making any bias in the timestamp ranges detectable and provable
 * Infrastructure performs commit-wait until the _client-selected_ `Timestamp`, not until server-generated `tt.Latest`, to ensure it will never give an overlapping `Timestamp-Range` to any causally-related request
 * Clients can pass around tickets or pass around records to detect or prove bias in `Timestamp-Range`s being handed out.
 * Clients can causally chain documents to prove causality violations in `Timestamp-Range`s.

### Fast/trusted mode

If a client trusts the infrastructure, the client can directly request a full notarization in one round trip by sending its data directly:

```http
X-Notary-Document-Data: <64 bytes base58 encoded>
```

The server then performs the above negotiation internally and returns the remaining fields to build a full notarization:

```http
X-Notary-Ticket-Timestamp-Range: <UTC nanosecond timestamp earliest>, <UTC nanosecond timestamp latest>
X-Notary-Ticket-Predecessor: <64 byte hash of "previous" ticket(s)>, ...
X-Notary-Ticket-Hash: <64 byte hash covering Timestamp-Range and Ticket-Predecessor(s) in order>
X-Notary-Document-Salt: <64 bytes base58 encoded>
X-Notary-Document-Timestamp: <UTC nanosecond timestamp within range of Timestamp-Range>
X-Notary-Document-Hash: <64 byte hash covering Ticket-Hash, Timestamp, Data, Salt>
X-Notary-Certificate-Predecessor: <64 byte hash of "previous" certificate(s)>, ...
X-Notary-Certificate-Hash: <64 byte hash covering Document-Hash, and Certificate-Predecessor(s) in order>
```

Note that in this case the server still generates (and logs) a ticket using the same format above. This ensures that notarizations obtained via the fast/trusted mode cannot be distinguished from those using paranoid mode above, and can still be part of any analysis of timestamp bias.

### Implementation

Internally, the notary generates a ticket using `tt.Earliest + 1*RTT` as the first timestamp and `tt.Latest + 3*RTT` for the second timestamp. Estimates for `RTT` can be a (large-ish) fixed value. If a client selects a later timestamp in this interval, the client's commit-wait will be longer. Clients will likely select the first timestamp.

### Open questions

 * Can clients re-order events since they have the power to choose the timestamps?
 * Earliest timestamp in the range has to be in the future. But causality?
 * Clients can collect tickets and redeem them in whatever order.
   * Shuold the infrastructure choose the ultimate timestamp as well, just pre-commit to the range?
   * Fail if ultimate timestamp cannot fit range? There could still be bias, but bias would be publishable.
> * Base-case attack: Client can get two tickets with the same range, create two causally chained documents (B=Hash(A)) and submit them in reverse timestamp order.
 * Delay attack: infrastructure publishes really big fixed ranges to everyone, biases itself as usual.
 * Delay attack: everyone may choose earliest timestamp, so delay attack morphs to a timestamp pre-commitment attack (which can be attacked by the client via ticket collection)
 * Can neither client nor infrastructure choose the timestamp?
   * Diffie-Hellman-ish negotiation?
   * Hash-based?
 * Diffie-Hellman negotiation to mix (timestamp_range, salt) with (user_data) and produce (timestamp, hash).

# Readings

 * [Ribbonfarm blog](https://www.ribbonfarm.com):
   * [After Temporality](https://www.ribbonfarm.com/2017/02/02/after-temporality/): the name fabula; the consensus timeline
   * [Markets Are Eating The World](https://www.ribbonfarm.com/2019/02/28/markets-are-eating-the-world/): mechanical clock towers providing "_fair_ (lower trust costs) and _fungible_ (lower transfer costs) measure of time"
