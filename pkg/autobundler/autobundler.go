// Package autobundler provides an automatic bundler for a stream of values.
//
// Autobundler tries to minimize the latency between when a value arrives and
// when it is handled. At steady state, autobundler will buffer incoming values
// while an invocation of the handler is running, and starting a new invocation
// with the next bundle of values as soon as the first invocation finishes.
package autobundler

import (
	"context"
	"reflect"
	"sync"
)

// AutoBundler manages
type AutoBundler struct {
	max        int
	valueCh    reflect.Value // chan T
	handler    func(context.Context, interface{})
	handlerCtx context.Context
	wg         sync.WaitGroup
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
// way to apply backpressure to upstream producers).
//
// The context ctx will be passed to the handler. If ctx is cancelled, no future
// handler invocations will occur.
func New(ctx context.Context, itemExample interface{}, handler func(ctx context.Context, v interface{}), max int) *AutoBundler {
	valueCh := reflect.MakeChan(reflect.ChanOf(reflect.BothDir, reflect.TypeOf(itemExample)), 0)
	r := &AutoBundler{
		max:        max,
		valueCh:    valueCh,
		handler:    handler,
		handlerCtx: ctx,
	}
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		nilSlice := reflect.Zero(reflect.SliceOf(reflect.TypeOf(itemExample)))
		accumBuf, handlerBuf := nilSlice, nilSlice
		var handlerCh chan struct{}
		handlerRunning := false
		for {
			casesWithSpace := []reflect.SelectCase{
				{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())},
				{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(handlerCh)},
				{Dir: reflect.SelectRecv, Chan: r.valueCh},
			}
			casesWithoutSpace := casesWithSpace[:2]
			var cases []reflect.SelectCase
			if accumBuf.Len() < max {
				cases = casesWithSpace
			} else {
				cases = casesWithoutSpace
			}
			chosen, val, _ := reflect.Select(cases)
			switch chosen {
			case 0:
				return
			case 1:
				handlerRunning = false
			case 2:
				accumBuf = reflect.Append(accumBuf, val)
			default:
				panic("select error")
			}
			if accumBuf.Len() > 0 && !handlerRunning {
				accumBuf, handlerBuf = handlerBuf, accumBuf
				accumBuf = accumBuf.Slice(0, 0)
				handlerCh = make(chan struct{})
				go func() {
					r.handler(ctx, handlerBuf.Interface())
					close(handlerCh)
				}()
				handlerRunning = true
			}
		}
	}()
	return r
}

// Add adds item to the current bundler.
//
// It is safe to call Add from multiple goroutines.
func (a *AutoBundler) Add(item interface{}) {
	a.valueCh.Send(reflect.ValueOf(item))
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
		panic("select error")
	}
}
