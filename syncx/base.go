package syncx

import (
	"errors"
	"io"
	"sync"
)

type Closes struct {
	mu      sync.Mutex
	closers []io.Closer
}

func (self *Closes) OnClosing(closers ...io.Closer) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.closers = append(self.closers, closers...)
}

func (self *Closes) CloseWith(closeHandle func() error) error {
	var err error
	if nil != closeHandle {
		err = closeHandle()
	}

	func() {
		self.mu.Lock()
		defer self.mu.Unlock()
		for _, closer := range self.closers {
			if e := closer.Close(); e != nil {
				if err == nil {
					err = e
				}
			}
		}
	}()
	return err
}

// type CloseW struct {
// 	C interface {
// 		Close()
// 	}
// }

// func (cw *CloseW) Close() error {
// 	if cw.C != nil {
// 		cw.C.Close()
// 	}
// 	return nil
// }

func ToCloser(c interface{}) io.Closer {
	if cw, ok := c.(interface {
		Close()
	}); ok {
		return CloseFunc(func() error {
			if cw != nil {
				cw.Close()
			}
			return nil
		})
	}

	if cf, ok := c.(func() error); ok {
		return CloseFunc(cf)
	}

	if closer, ok := c.(io.Closer); ok {
		return closer
	}
	panic(errors.New("it isn't a closer"))
}

type CloseFunc func() error

func (f CloseFunc) Close() error {
	if f == nil {
		return nil
	}
	return f()
}
