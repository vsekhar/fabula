# TrueTime

Most clocks tell you what time it _is_. TrueTime tells you what time it _isn't_.

In response to a `now()` request, TrueTime provides two times: `tt.earliest` and `tt.latest`. TrueTime semantics state that the time during the call to `now()` was _not_ before `tt.earliest` and was _not_ after `tt.latest`.

It turns out this odd formulation is exceedingly powerful for managing concurrency in distributed systems.

## Resources

TrueTime was first described briefly in the [Spanner paper](https://ai.google/research/pubs/pub39966) and more generally in the [documentation for Google's Cloud Spanner service](https://cloud.google.com/spanner/docs/true-time-external-consistency).

