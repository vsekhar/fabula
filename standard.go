// Package COMMIT defines constants and strings for the COMMIT HTTP verb.
package COMMIT

// Header names
const (
	EarliestHeader     = "Consistent-Earliest"
	LatestHeader       = "Consistent-Latest"
	EpsilonDebugHeader = "x-Consistent-Epsilon"
	TimestampHeader    = "Consistent-Timestamp"
)

// Paths
const (
	TrueTimeNowPath        = "/.well-known/truetime/now"
	TrueTimeCommitWaitPath = "/.well-known/truetime/commitwait"
)
