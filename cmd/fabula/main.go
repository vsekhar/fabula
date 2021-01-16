package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// Notarize data from a file (hashes the file then notarizes the hash) and
// outputs a notarization hash, timestamp (and optionally the icnlusion proof):
//
//   $ fabula notarize --data_file=<file> [--with_proof]
//
// Notarize 64 bytes of data on the command line:
//
//   $ fabula notarize --data_base64=<data> [--with_proof]
//
// Data is of the form "<URL safe base64 encoding>".
//
// Prove inclusion of an entry to its sequence (batch proof) and display
// sequence entry and timestamp:
//
//   $ fabula prove --notarization=<hash>
//
// Hashes are of the form "sha3512:<64 bytes URL safe base64 encoding>".
//
// Prove inclusion of an entry to its sequence (batch proof) and consistency of
// its sequence to some future sequence, emit to standard output (typically
// written to a file or piped to verfy):
//
//   $ fabula prove --notarization=<hash> --recent > <file>
//   $ fabula prove --notarization=<hash> --end=<time> > <file>
//
// Consistency between two sequences (can be appended to output of `prove` to
// "evergreen" the proof):
//
//   $ fabula consistency --start=<time> --end=<time> >> <file>
//
// Times are of the form "<UTC nanoseconds since the Unix epoch>".
//
// Verify an inclusion proof, reading from a file (or `-` for stdin), or from
// the global store or a proxy:
//
//   $ fabula verify --data_base64=<data> --salt=<salt> --end=<time> <file>
//   $ fabula verify --notarization=<hash> --end=<time> <file>
//
// Verify consistency, reading from a file (or `-` for stdin):
//
//   $ fabula verify --start=<time> --end=<time> <file>

type stringsFlag []string

func (s *stringsFlag) String() string         { return fmt.Sprintf("%v", *s) }
func (s *stringsFlag) Set(value string) error { *s = append(*s, value); return nil }

func main() {
	var (
		notarizeCommand   = flag.NewFlagSet("notarize", flag.ExitOnError)
		notarizeDataFile  = notarizeCommand.String("data_file", "", "file to notarize")
		notarizeData      = notarizeCommand.String("data_base64", "", "data to notarize, encoded as a base64 string")
		notarizeWithProof = notarizeCommand.Bool("with_proof", false, "output the inclusion proof of the notarization")
	)

	var (
		proveCommand      = flag.NewFlagSet("prove", flag.ExitOnError)
		proveNotarization = proveCommand.String("notarization", "", "notarization hash to prove")
		proveRecent       = proveCommand.Bool("recent", false, "prove to some recent timestamp")
		proveEndTime      = proveCommand.Uint64("end", 0, "end time for proof")
	)

	var (
		consistencyCommand   = flag.NewFlagSet("consistency", flag.ExitOnError)
		consistencyStartTime = consistencyCommand.Uint64("start", 0, "start time for consistency")
		consistencyEndTime   = consistencyCommand.Uint64("end", 0, "end time for consistency")
	)

	var (
		verifyCommand      = flag.NewFlagSet("verify", flag.ExitOnError)
		verifyData         = verifyCommand.String("data_base64", "", "data to verify")
		verifySalt         = verifyCommand.String("salt", "", "salt for data being verified")
		verifyNotarization = verifyCommand.String("notarization", "", "notarization hash to verify")
		verifyEndTime      = verifyCommand.Uint64("end", 0, "end time for verification")
	)

	flagSets := map[string]*flag.FlagSet{
		"notarize":    notarizeCommand,
		"prove":       proveCommand,
		"consistency": consistencyCommand,
		"verify":      verifyCommand,
	}

	// common flags
	var cachePath string
	var maxCacheSizeMb uint
	var format string
	proxies := &stringsFlag{}
	for _, f := range flagSets {
		f.StringVar(&cachePath, "cache_path", "", "path to cache directory (default: system-defined cache dir")
		f.UintVar(&maxCacheSizeMb, "cache_max", 100, "maximum cache size in MB")
		f.Var(proxies, "proxy", "proxy to consult; multiple may be provided; when looking up batches and sequences, proxies will be consulted in the order provided")
		f.StringVar(&format, "format", "text", "format for output: 'text', 'proto'")
	}

	if len(os.Args) < 2 {
		fmt.Print("command required: notarize, prove, consistency, or verify.")
		os.Exit(1)
	}

	if f, ok := flagSets[os.Args[1]]; ok {
		if err := f.Parse(os.Args[2:]); err != nil {
			fmt.Print(err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("bad command: %s", os.Args[1])
	}

	if cachePath == "" {
		userCacheDir, err := os.UserCacheDir()
		if err != nil {
			fmt.Printf("no cache_path provided and could not get default user cachedir: %v", err)
			os.Exit(1)
		}
		cachePath = filepath.Join(userCacheDir, "fabula")
	}

	if notarizeCommand.Parsed() {
		_ = notarizeDataFile
		_ = notarizeData
		_ = notarizeWithProof
		panic("unimplemented")
	}

	if proveCommand.Parsed() {
		_ = proveNotarization
		_ = proveRecent
		_ = proveEndTime
		panic("unimplemented")
	}

	if consistencyCommand.Parsed() {
		_ = consistencyStartTime
		_ = consistencyEndTime
		panic("unimplemented")
	}

	if verifyCommand.Parsed() {
		_ = verifyData
		_ = verifySalt
		_ = verifyNotarization
		_ = verifyEndTime
		panic("unimplemented")
	}
}
