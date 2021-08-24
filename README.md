# Fabula

_Globally consistent verifiable private consensus at 1 million events per second._

> _Fabula (Russian: фабула, IPA: [ˈfabʊlə]) is the chronological order of the events contained in a story, the "raw material" of a story._ -[Wikipedia](https://en.wikipedia.org/wiki/Fabula_and_syuzhet)

Fabula is a proof-of-concept stack of services and protocols for verifiable decentralized action on the web.

## Rationale

Consensus and synchrony are not independent ideas.

Consensus is not a state, it is an event. It is tied to the circumstances of its moment. It can occur from nothing and dissipate. Consensus is ephemeral as the participants and the world around them evolve.

For consensus to have any lasting effect, it must be immortalized. As a result, the goal of an algorithm can be properly understood as the communal creation of a single immortalization.

An immortalization exists as a record of a past event. While immortalizations may each have their own distinct semantics, they all shere this _pastness_ property. Immortalizations can thus be compared with each other on the basis of their pastness. These comparisons yield a partial _ordering_ of immortalizations.

As a result, to agree on a fact, participants also need to agree on an ordering on which that consensus will be immortalized.

## Readings

* TBD: blog posts
* [Ribbonfarm blog](https://www.ribbonfarm.com):
  * [After Temporality](https://www.ribbonfarm.com/2017/02/02/after-temporality/): the name fabula; the consensus timeline
  * [Markets Are Eating The World](https://www.ribbonfarm.com/2019/02/28/markets-are-eating-the-world/): mechanical clock towers providing "_fair_ (lower trust costs) and _fungible_ (lower transfer costs) measure of time"
* Companies working on aspects of time:
  * [Building a more accurate time service at Facebook scale, Mar. 2020.](https://engineering.fb.com/production-engineering/ntp-service/)
  * [Keeping Time With Amazon Time Sync Service, Nov. 2017.](https://aws.amazon.com/blogs/aws/keeping-time-with-amazon-time-sync-service/)
  * [Cloud Spanner: TrueTime and external consistency](https://cloud.google.com/spanner/docs/true-time-external-consistency)
    * [Spanner, TrueTime and the CAP Theorem](https://research.google/pubs/pub45855/)
    * [Spanner: Google's Globally-Distributed Database](https://research.google/pubs/pub39966/)
