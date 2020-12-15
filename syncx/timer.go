package syncx

import (
	"errors"
	"io"
	"sync/atomic"
	"time"
)

type Timer struct {
	isRunning int32
	interval  time.Duration
	timer     atomic.Value
	cb        func() bool
}

func (timer *Timer) Interval() time.Duration {
	return timer.interval
}

func (timer *Timer) Start(interval time.Duration, cb func() bool) {
	if cb == nil {
		panic(errors.New("argument 'cb' is nil!"))
	}
	if timer.timer.Load() != nil {
		panic(errors.New("timer is initialized!"))
	}
	timer.cb = cb
	timer.interval = interval

	newtimer := time.AfterFunc(timer.interval, timer.tick)
	timer.timer.Store(newtimer)
}

func (timer *Timer) Stop() {
	if o := timer.timer.Load(); o != nil {
		if t := o.(*time.Timer); t != nil {
			t.Stop()
		}
	}
}

func (timer *Timer) tick() {
	if !atomic.CompareAndSwapInt32(&timer.isRunning, 0, 1) {
		return
	}

	defer atomic.StoreInt32(&timer.isRunning, 0)

	if timer.cb() {
		if o := timer.timer.Load(); o != nil {
			if t := o.(*time.Timer); t != nil {
				t.Reset(timer.interval)
				return
			}
		}

		newtimer := time.AfterFunc(timer.interval, timer.tick)
		timer.timer.Store(newtimer)
	}
}

type Tickable struct {
	isClosed  int32
	isRunning int32
	interval  time.Duration
	timer     atomic.Value
	cb        func()
	closes    Closes
}

func (tk *Tickable) Init(interval time.Duration, cb func()) {
	if cb == nil {
		panic("cb is nil!")
	}
	if tk.cb != nil {
		panic("Tickable is initialized!")
	}
	tk.cb = cb
	tk.interval = interval

	cb()

	timer := time.AfterFunc(tk.interval, tk.tick)
	tk.timer.Store(timer)
}

func (tk *Tickable) OnClosing(closers ...io.Closer) {
	tk.closes.OnClosing(closers...)
}

func (tk *Tickable) IsClosed() bool {
	return atomic.LoadInt32(&tk.isClosed) != 0
}

func (tk *Tickable) Close() error {
	if atomic.CompareAndSwapInt32(&tk.isClosed, 0, 1) {
		return tk.closes.CloseWith(func() error {
			if o := tk.timer.Load(); o != nil {
				if t := o.(*time.Timer); t != nil {
					t.Stop()
				}
			}
			return nil
		})
	}
	return nil
}

func (tk *Tickable) tick() {
	if tk.IsClosed() {
		return
	}

	if !atomic.CompareAndSwapInt32(&tk.isRunning, 0, 1) {
		return
	}

	defer func() {
		atomic.StoreInt32(&tk.isRunning, 0)

		if tk.IsClosed() {
			return
		}
		if o := tk.timer.Load(); o != nil {
			if t := o.(*time.Timer); t != nil {
				t.Reset(tk.interval)
				return
			}
		}

		timer := time.AfterFunc(tk.interval, tk.tick)
		tk.timer.Store(timer)
	}()

	tk.cb()
}

func Sleep(timeout time.Duration, s <-chan struct{}) {
	timer := time.NewTimer(timeout)
	select {
	case <-timer.C:
	case <-s:
		timer.Stop()
	}
}
