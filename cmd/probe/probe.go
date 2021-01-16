// probe fetches and prints probes for statistical analysis.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/vsekhar/fabula/internal/youtime"
)

var n = flag.Int("n", 0, "maximum number of probes to fetch (0 is unlimited)")
var host = flag.String("h", "time.google.com", "NTP host")
var port = flag.Int("p", 123, "NTP port")
var outFile = flag.String("o", "", "output file (or stdout if empty)")
var interval = flag.Duration("i", 1*time.Second, "gap between probes")

func main() {
	flag.Parse()
	var out *os.File
	if *outFile != "" {
		var err error
		out, err = os.Create(*outFile)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		out = os.Stdout
	}
	if out != os.Stdout {
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC | log.Lshortfile)
	}

	nc, err := net.Dial("udp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		log.Fatal(err)
	}

	var bar *progressbar.ProgressBar
	if out != os.Stdout {
		bar = progressbar.Default(-1, "Probing...")
	}
	fmt.Fprint(out, "Local, Out-upper, In-lower, Out-upper-impure, In-lower-impure, Out-jitter, In-jitter\n")
	for i := 0; *n == 0 || i < *n; i++ {
		op, ip, err := youtime.GetCodedProbes(nc)
		if err != nil && err != youtime.ErrCodedProbesNotPure {
			log.Fatal(err)
		}

		// We use the local time of op for both the upper and lower bounds. This
		// skews lower bounds left by ~1RTT. This is ok for graphing purposes
		// and lets us record jitter for each probe pair on the same row. The
		// real SVM would use the correct local time for each probe.

		format := "%d, %d, %d, , , %d, %d\n"
		if err == youtime.ErrCodedProbesNotPure {
			format = "%d, , , %d, %d, %d, %d\n"
		}
		fmt.Fprintf(out, format,
			youtime.DebugLocal(op[0]),
			youtime.Bound(op[0]),
			youtime.Bound(ip[0]),
			youtime.Jitter(op[0], op[1]).Nanoseconds(),
			youtime.Jitter(ip[0], ip[1]).Nanoseconds(),
		)
		fmt.Fprintf(out, format,
			youtime.DebugLocal(op[1]),
			youtime.Bound(op[1]),
			youtime.Bound(ip[1]),
			youtime.Jitter(op[0], op[1]).Nanoseconds(),
			youtime.Jitter(ip[0], ip[1]).Nanoseconds(),
		)
		if out != os.Stdout {
			bar.Add(1)
		}
		time.Sleep(1 * time.Second)
	}
}
