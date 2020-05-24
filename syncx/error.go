package syncx

import (
	"io"
	"sync/atomic"
)

type errorWrapper struct {
	err error
}

type ErrorValue struct {
	value atomic.Value
}

func (ev *ErrorValue) Set(e error) {
	ev.value.Store(&errorWrapper{err: e})
}

func (ev *ErrorValue) Get() error {
	o := ev.value.Load()
	if o == nil {
		return nil
	}
	if e, ok := o.(*errorWrapper); ok {
		return e.err
	}
	return nil
}

type closeWrapper struct {
	v io.Closer
}

func (cw *closeWrapper) Close() error {
	if cw.v == nil {
		return nil
	}
	return cw.v.Close()
}

type CloseWrapper struct {
	v atomic.Value
}

func (cw *CloseWrapper) Set(closer io.Closer) {
	cw.v.Store(&closeWrapper{v: closer})
}

func (cw *CloseWrapper) Close() error {
	o := cw.v.Load()
	if o == nil {
		return nil
	}
	if closer, ok := o.(io.Closer); ok && closer != nil {
		err := closer.Close()
		cw.Set(nil)
		return err
	}
	return nil
}
