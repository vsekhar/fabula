# TrueTime

Most clocks tell you what time it _is_. TrueTime tells you what time it _isn't_.

In response to a `now()` request, TrueTime provides two times: `tt.earliest` and `tt.latest`. Informally, you can think of TrueTime providing a bounded guess at the time.

More formally, TrueTime asserts that there existed a moment after the call to `now()` commenced but before it returned when the time was _not_ before `tt.earliest` and was _not_ after `tt.latest`.

This confusing formulation isn't helpful for _timing_ (ironically), but is essential for _ordering_. And ordering is essential for building efficient and consistent distributed systems.00

Read more about TrueTime in the [Spanner paper](https://ai.google/research/pubs/pub39966) and the [documentation for Google's Cloud Spanner service](https://cloud.google.com/spanner/docs/true-time-external-consistency).

