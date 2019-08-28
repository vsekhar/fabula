// Package commit defines constants and strings for the COMMIT HTTP verb.
package commit

// Header names
const (
	EarliestHeader     = "Consistent-Earliest"
	LatestHeader       = "Consistent-Latest"
	TimestampHeader    = "Consistent-Timestamp"
	EpsilonDebugHeader = "x-Consistent-Epsilon"
	FakeDebugHeader    = "x-Consistent-Fake"
)

// Paths
const (
	TrueTimeNowPath        = "/.well-known/truetime/now"
	TrueTimeCommitWaitPath = "/.well-known/truetime/commitwait"
)
