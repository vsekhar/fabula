package notify

// Notifier can be used to trigger a callback in a coalescing way. That is,
// the callback will be called only once to cover all calls to Notify() that
// were made while the last iteration of the callback was running.
//
// There will only be one call to callback in flight at a time.
type Notifier struct {
	ch       chan struct{}
	callback func()
}

// New returns a new Notifier.
func New(callback func()) *Notifier {
	r := &Notifier{
		ch:       make(chan struct{}, 1), // must have buffer size 1
		callback: callback,
	}
	go func() {
		for range r.ch {
			if r.callback != nil {
				r.callback()
			}
		}
	}()
	return r
}

// Notify notifies the callback. If the callback is currently running, all calls
// to Notify are coalesced to a single subsequent call.
func (n *Notifier) Notify() {
	select {
	case n.ch <- struct{}{}:
	default:
	}
}

// Close closes the Notifier and releases its resources.
func (n *Notifier) Close() {
	close(n.ch)
}
