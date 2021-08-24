# GCP Locations

| Region (phase)    | ID                   | Cloud Run     | Cloud Storage | Cloud KMS      | GCE   |
|-------------------|----------------------|---------------|---------------|----------------|-------|
| **US Iowa (1)**   | us-central1          | Tier 1        | Base          | Software & HSM | Base  |
| **Belgium (1)**   | europe-west1         | Tier 1        | Base          | Software & HSM | +10%  |
| Zurich            | europe-west6         | N/A           | +25%          | Software & HSM | +40%  |
| Tokyo             | asia-northeast1      | Tier 1        | +15%          | Software & HSM | +29%  |
| **Osaka (2)**     | asia-northeast2      | Tier 1        | +15%          | Software & HSM | +29%  |
| **Singapore (1)** | asia-southeast1      | Tier 2 (+40%) | Base          | Software & HSM | +23%  |
| **Sydney (2)**    | australia-southeast1 | Tier 2 (+40%) | +15%          | Software & HSM | +42%  |
| Mumbai            | asia-south1          | N/A           | +15%          | Software & HSM | +20%  |

* Netherlands (europe-west4) is equivalent to Belgium for the above services, but is newer and has slightly higher GCE pricing, suggesting a smaller data center presence.
* Osaka is slightly closer to the rest of Asia and has more transpacific cables (at Shima) than Tokyo.
* Zurich (non-EU) and Mumbai (South and Central Asia) can be accessed via GCE rather than Cloud Run.

## Cloud Storage

Optimize for writes.

* Regional: fastest and cheapest
* Dual-region: same performance, higher availability, higher cost
* **Multi-region**: similar and sometimes faster performance (!), higher cost

NB: see `storage/` for benchmarks.

| Class     | Rest $/gb-mth | Writes $/10k | Reads $/10k |
|-----------|---------------|--------------|-------------|
| Standard  | $0.020        | $0.05        | $0.004      |
| Nearline  | $0.010        | $0.10        | $0.01       |
| Coldline  | $0.004        | $0.10        | $0.05       |
| Archive   | $0.0012       | $0.50        | $0.50       |

Can use Object Lifecycle rules to change the storage class of objects after a certain age.

Old objects are summarized by newer objects so are no longer needed for summaries. Old objects would only be read for entry-specific proofs.

Can pack objects and do range reads to read many entries with a single read op. MMR-aware packing could pack at major peaks and label the pack by the index of the peak. Packs can be self-verified, without referring back to the original MMR: a spurious error is unlikely to produce a valid pack.

## Avoid

* asia-east1: Taiwan
* asia-east2: Hong Kong
