package shell

import (
	"io"
	"sync"
	"sync/atomic"
	"time"
)

type drainto interface {
	DrainTo(w io.Writer) (int, error)
}

var _ drainto = &pipe{}

var bytecache = sync.Pool{
	New: func() interface{} {
		return make([]byte, 1024)
	},
}

type pipe struct {
	c                         chan byte
	isClosed                  int32
	readTimeout, writeTimeout time.Duration
}

func (c *pipe) SetReadDeadline(t time.Duration) error {
	c.readTimeout = t
	return nil //c.rc.SetTimeout(t)
}

func (c *pipe) SetWriteDeadline(t time.Duration) error {
	c.writeTimeout = t
	return nil //errors.New("notimplented")
}

func (c *pipe) Close() error {
	if atomic.CompareAndSwapInt32(&c.isClosed, 0, 1) {
		close(c.c)
	}
	return nil
}

func (c *pipe) IsClosed() bool {
	return atomic.LoadInt32(&c.isClosed) != 0
}

func (c *pipe) WriteByte(b byte) (err error) {
	if c.IsClosed() {
		return io.EOF
	}

	// fmt.Println("pipe write:", string(p))

	defer func() {
		if o := recover(); o != nil {
			err = io.ErrClosedPipe
		}
	}()

	if c.writeTimeout <= 0 {
		c.c <- b
		return nil
	}

	timer := time.NewTimer(c.writeTimeout)
	select {
	case c.c <- b:
		timer.Stop()
		return nil
	case <-timer.C:
		return io.ErrShortWrite
	}
}

func (c *pipe) Write(p []byte) (n int, err error) {
	if c.IsClosed() {
		return 0, io.EOF
	}

	// fmt.Println("pipe write:", string(p))

	var timer *time.Timer
	if c.writeTimeout > 0 {
		timer = time.NewTimer(c.writeTimeout)
	}
	defer func() {
		if o := recover(); o != nil {
			err = io.ErrClosedPipe
		}
		if timer != nil {
			timer.Stop()
		}
	}()
	for idx := range p {
		if timer != nil {
			select {
			case c.c <- p[idx]:
				n = idx + 1
			case <-timer.C:
				err = io.ErrShortWrite
				return
			}
		} else {
			c.c <- p[idx]
			n = idx + 1
		}
	}
	return
}

func (c *pipe) Read(p []byte) (int, error) {
	if c.readTimeout > 0 {
		timer := time.NewTimer(c.readTimeout)
		offset := 0
		for {
			select {
			case b, ok := <-c.c:
				if !ok {

					timer.Stop()
					return offset, io.EOF
				}
				// fmt.Println("pipe read:", string(b))

				p[offset] = b
				offset++
				if len(p) <= offset {

					timer.Stop()
					return offset, nil
				}
			case <-timer.C:
				return offset, nil
			}
		}
	} else {
		offset := 0
		for {
			select {
			case b, ok := <-c.c:
				if !ok {
					return offset, io.EOF
				}
				// fmt.Println("pipe read:", string(b))

				p[offset] = b
				offset++
				if len(p) <= offset {
					return offset, nil
				}
			default:
				return offset, nil
			}
		}
	}
}

func (c *pipe) DrainTo(w io.Writer) (int, error) {
	var a [1]byte
	offset := 0
	for {
		select {
		case b, ok := <-c.c:
			if !ok {
				return offset, io.EOF
			}
			offset++
			a[0] = b

			w.Write(a[:])

			// fmt.Println("pipe read:", string(b))
		default:
			return offset, nil
		}
	}
}

func (c *pipe) ReadByte() (byte, error) {
	if c.readTimeout > 0 {
		timer := time.NewTimer(c.readTimeout)
		select {
		case b, ok := <-c.c:
			timer.Stop()

			if !ok {
				return 0, io.EOF
			}

			// fmt.Println("pipe read:", string(b))
			return b, nil
		case <-timer.C:
			return 0, ErrTimeout
		}
	}
	b, ok := <-c.c
	if !ok {
		return 0, io.EOF
	}
	// fmt.Println("pipe read:", string(b))
	return b, nil
}

func MakePipe(timeout time.Duration) *pipe {
	return &pipe{c: make(chan byte, 2048)}
}
