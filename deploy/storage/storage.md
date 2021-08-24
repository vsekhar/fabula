# Storage format

| Region                    | Median updates @ 1KB | Metadata @ 1KB |
| --------------------------|----------------------|----------------|
| Multi-region US           | 257.3 ms             | 63.3 ms        |
| Dual-region NAM4          | 314.5 ms             | 71.4 ms        |
| Single-region USCENTRAL-1 | 176.7 ms             | 71.1 ms        |

Probably want multi-region. Could be really fast if everything is in metadata...
