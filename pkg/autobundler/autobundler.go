// Package autobundler provides an automatic bundler for a stream of values.
//
// Unlike `google.golang.org/api/support/bundler`, package autobundler does not
// require setting timeouts or thresholds.
//
// Instead, Autobundler tries to minimize the latency between when a value
// arrives and when it is handled. At steady state, autobundler will buffer
// incoming values while an invocation of the handler is running, starting a new
// invocation with the next bundle of values as soon as the prior invocation
// finishes.
package autobundler

import (
	"context"
	"reflect"
	"sync"
)

// AutoBundler bundles values.
//
// The methods of AutoBundler are safe to call from multiple goroutines.
type AutoBundler struct {
	max     int
	valueCh reflect.Value // chan T
	ctx     context.Context
	wg      sync.WaitGroup
}

// New returns a new AutoBundler.
//
// The itemExample argument is used to set the type of values the AutoBundler
// will handle.
//
// Handler is called and passed a bundle of values to be handled. Only one
// instance of handler will be running at a time. Values are passed to handler
// in the order they were added to the AutoBundler.
//
// If itemExample is of type T, then handler will be passed an argument of type
// []T. For example, if you want to create bundles of type `*Entry` then you can
// pass `&Entry{}` as itemExample and handler will then be passed an argument
// of type `[]*Entry`.
//
// The AutoBundler will only buffer max values, after which calls to Add will
// block until the bundle is submitted to the handler. Call AddNoWait if you
// want to detect if the buffer is full. Setting a reasonable max provides a
// way to apply backpressure to upstream producers. For handlers that are not
// CPU constrained (e.g. which write bundles to remote storage), a max of 1000
// is reasonable.
//
// The context ctx will be passed to the handler. If ctx is cancelled, no future
// handler invocations will occur and the autobundler will stop handling
// buffered and new values. There is no way to safely "flush" an autobundler.
// Applications should be resilient to losing values buffered in the autobundler
// if the context is cancelled. In other words, context cancellation is treated
// like termination of the program from the point of view of the autobundler.
func New(ctx context.Context, itemExample interface{}, handler func(ctx context.Context, v interface{}), max int) *AutoBundler {
	valueCh := reflect.MakeChan(reflect.ChanOf(reflect.BothDir, reflect.TypeOf(itemExample)), 0)
	r := &AutoBundler{
		max:     max,
		valueCh: valueCh,
		ctx:     ctx,
	}
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		typ := reflect.TypeOf(itemExample)
		nilSlice := reflect.Zero(reflect.SliceOf(typ))
		buf := nilSlice
		var handlerCh chan struct{}
		handlerRunning := false
		allCases := []reflect.SelectCase{
			{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())},
			{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(handlerCh)},
			{Dir: reflect.SelectRecv, Chan: r.valueCh}, // add new value
		}
		for {
			var chosen int
			var val reflect.Value
			cases := allCases
			if buf.Len() >= max {
				// Exclude receive case if the buffer is full. This will cause
				// calls to Add to block, and calls to AddNoWait to return false.
				cases = cases[:2]
			}
			chosen, val, _ = reflect.Select(cases)
			switch chosen {
			case 0:
				// <-ctx.Done()
				if handlerRunning {
					<-handlerCh
				}
				return
			case 1:
				// <-handlerCh
				handlerRunning = false
			case 2:
				// val = <-valueCh
				buf = reflect.Append(buf, val)
			default:
				panic("autobundler.New: select error")
			}

			if buf.Len() > 0 && !handlerRunning {
				handlerBuf := buf
				// pre-allocate length of last slice
				buf = reflect.MakeSlice(reflect.SliceOf(typ), 0, handlerBuf.Len())
				handlerCh = make(chan struct{})
				cases[1].Chan = reflect.ValueOf(handlerCh)
				go func() {
					handler(ctx, handlerBuf.Interface())
					close(handlerCh)
				}()
				handlerRunning = true
			}
		}
	}()
	return r
}

// Add adds item to the current bundler, or returns an error.
//
// It is safe to call Add from multiple goroutines.
func (a *AutoBundler) Add(ctx context.Context, item interface{}) error {
	cases := []reflect.SelectCase{
		{Dir: reflect.SelectSend, Chan: a.valueCh, Send: reflect.ValueOf(item)},
		{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(a.ctx.Done())},
		{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())},
	}
	chosen, _, _ := reflect.Select(cases)
	switch chosen {
	case 0:
	case 1:
		return a.ctx.Err()
	case 2:
		return ctx.Err()
	default:
		panic("autobundler.Add: select error")
	}
	return nil
}

// AddNoWait tries to add an item to the current bundler. If successful, it
// returns true, otherwise it returns false. AddNoWait does not block.
//
// It is safe to call AddNoWait from multiple goroutines.
func (a *AutoBundler) AddNoWait(item interface{}) bool {
	cases := []reflect.SelectCase{
		{Dir: reflect.SelectSend, Chan: a.valueCh, Send: reflect.ValueOf(item)},
		{Dir: reflect.SelectDefault},
	}
	chosen, _, _ := reflect.Select(cases)
	switch chosen {
	case 0:
		return true
	case 1:
		return false
	default:
		panic("autobundler.AddNoWait: select error")
	}
}

// Wait blocks until the autobundler has stopped. The autobundler will stop when
// the context passed to New is cancelled.
func (a *AutoBundler) Wait() {
	a.wg.Wait()
}
