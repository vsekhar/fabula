package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	ansi "github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
)

var enc = hex.EncodeToString
var dec = hex.DecodeString

func pbopts() []progressbar.Option {
	w := ansi.NewAnsiStdout()
	if *verbose {
		w = ioutil.Discard
	}
	return []progressbar.Option{
		progressbar.OptionSetWriter(w),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetItsString("entries"),
		progressbar.OptionShowIts(),
		progressbar.OptionShowCount(),
		progressbar.OptionThrottle(100 * time.Millisecond),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	}
}

func doPar(ctx context.Context, f func(context.Context)) {
	wg := sync.WaitGroup{}
	wg.Add(*p)
	ch := make(chan struct{}, *p)
	for i := 0; i < *p; i++ {
		go func() {
			for range ch {
				f(ctx)
			}
			wg.Done()
		}()
	}
	for i := 0; i < *n; i++ {
		ch <- struct{}{}
	}
	close(ch)
	wg.Wait()
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC | log.Lshortfile)
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	if *verbose {
		log.Println("starting")
		log.Println("tasks:", *p)
	}
	if *n > 0 && *verbose {
		log.Println("n:", *n)
	}
	if *timeout != time.Duration(0) {
		if *verbose {
			log.Printf("timeout: %s\n", *timeout)
		}
		ctx, cancel = context.WithTimeout(ctx, *timeout)
	}
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, syscall.SIGTERM)
	go func() { <-s; cancel() }()

	initMonitoring(ctx)
	initTrace(ctx)

	args := flag.Args()
	if len(args) != 1 {
		fmt.Println("missing command: {create, load, drain}")
		fmt.Println("args:", args)
		flag.PrintDefaults()
		os.Exit(1)
	}

	switch args[0] {
	case "create":
		create(ctx)
	}

	switch args[0] {
	case "load":
		load(ctx)
	case "drain":
		drain(ctx)
	case "pack":
		pack(ctx)
	default:
		fmt.Println("Unknown command", args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *verbose {
		log.Println("done")
	}
}
