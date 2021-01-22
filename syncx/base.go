package syncx

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
)

type Base struct {
	closed int32
	S      chan struct{}
	Wait   sync.WaitGroup

	Closes
}

func (base *Base) CloseWith(closeHandle func() error) error {
	if !atomic.CompareAndSwapInt32(&base.closed, 0, 1) {
		return nil
	}
	err := base.Closes.CloseWith(func() error {
		if nil != base.S {
			close(base.S)
		}

		if nil != closeHandle {
			return closeHandle()
		}
		return nil
	})
	base.Wait.Wait()
	return err
}

func (base *Base) IsClosed() bool {
	return 0 != atomic.LoadInt32(&base.closed)
}

func (base *Base) CatchThrow(message string, err *error) {
	if o := recover(); nil != o {
		var buffer bytes.Buffer
		if "" == message {
			buffer.WriteString(fmt.Sprintf("[panic] %v", o))
		} else {
			buffer.WriteString(fmt.Sprintf("[panic] %v - %v", message, o))
		}

		for i := 1; ; i++ {
			pc, file, line, ok := runtime.Caller(i)
			if !ok {
				break
			}
			funcinfo := runtime.FuncForPC(pc)
			if nil != funcinfo {
				buffer.WriteString(fmt.Sprintf("    %s:%d %s\r\n", file, line, funcinfo.Name()))
			} else {
				buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
			}
		}

		errMsg := buffer.String()
		log.Println(errMsg)
		if nil != err {
			*err = errors.New(errMsg)
		}
	}
}

func (base *Base) RunItInGoroutine(cb func()) {
	base.Wait.Add(1)
	go func() {
		cb()
		base.Wait.Done()
	}()
}
