package util

import (
	"errors"
	"sync/atomic"
	"time"
)

type Timer struct {
	isRunning int32
	interval  time.Duration
	timer     atomic.Value
	cb        func() bool
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
