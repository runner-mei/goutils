package syncx

import (
	"sync/atomic"
)

type Threads struct {
	c          chan func()
	threads    int32
	maxThreads int32
}

func (pths *Threads) Run(cb func()) bool {
	pths.c <- cb

	for i := 0; i < 10; i++ {
		threads := atomic.LoadInt32(&pths.threads)
		if threads > pths.maxThreads {
			return false
		}

		if atomic.CompareAndSwapInt32(&pths.threads, threads, threads+1) {
			go func() {
				defer atomic.AddInt32(&pths.threads, -1)

				for {
					select {
					case f, ok := <-pths.c:
						if !ok {
							return
						}

						f()
					default:
						return
					}
				}
			}()
			return true
		}
	}
	return false
}

func NewThreads(queueSize, maxThreads int) *Threads {
	if maxThreads <= 0 {
		maxThreads = 2
	}

	return &Threads{
		c:          make(chan func(), queueSize),
		maxThreads: int32(maxThreads),
	}
}
