package interrupt

import (
	"os"
	"os/signal"
)

// Wait blocks until an interrupt has been received.
//
// Wait is usually called at the bottom of main() in a long-running process or
// server.
func Wait() {
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh
}
