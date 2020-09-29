package batching_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/vsekhar/fabula/internal/batching"
)

func TestBatcher(t *testing.T) {
	items := make(chan interface{})
	batches := batching.Channel(items, 2, 1+time.Second)
	go func() {
		for i := 0; i < 9; i++ {
			items <- i
		}
		time.Sleep(1 * time.Second)
		close(items)
	}()
	out := strings.Builder{}
	for b := range batches {
		out.WriteString(fmt.Sprintf("%v,", b))
	}
	expected := "[0 1],[2 3],[4 5],[6 7],[8],"
	if out.String() != expected {
		t.Errorf("expected '%s', got '%s'", expected, out.String())
	}
}
