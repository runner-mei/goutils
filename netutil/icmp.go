package netutil

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/runner-mei/errors"
)

var logger = log.New(os.Stderr, "[ping] ", log.Lshortfile|log.LstdFlags)

var resultArrayPool = sync.Pool{New: func() interface{} {
	return make([]*PingResult, 0, 1000)
}}

var resultPool = sync.Pool{New: func() interface{} {
	return &PingResult{}
}}

const (
	ICMP4_ECHO_REQUEST = 8
	ICMP4_ECHO_REPLY   = 0
	ICMP6_ECHO_REQUEST = 128
	ICMP6_ECHO_REPLY   = 129
)

type PingResult struct {
	Addr      net.Addr
	Id        uint16
	SeqNum    uint16
	Bytes     []byte
	Err       error
	Timestamp time.Time
}

func (self *PingResult) SetRequestID(id uint32) {
	self.Id = uint16(id >> 16)
	self.SeqNum = uint16(id & 0xffff)
}

func (self *PingResult) GetRequestID() uint32 {
	id := uint32(self.Id) << 16
	id = id | uint32(self.SeqNum)
	return id
}

func Uint16ToUint32(high, low uint16) uint32 {
	id := uint32(high) << 16
	id = id | uint32(low)
	return id
}

func Uint32ToUint16(i uint32) (uint16, uint16) {
	return uint16(i >> 16), uint16(i & 0xffff)
}

func family(a *net.IPAddr) int {
	if a == nil || len(a.IP) <= net.IPv4len {
		return syscall.AF_INET
	}
	if a.IP.To4() != nil {
		return syscall.AF_INET
	}
	return syscall.AF_INET6
}

type Pinger struct {
	family       int
	network      string
	seqnum       uint16
	id           uint16
	echo         []byte
	conn         net.PacketConn
	wait         sync.WaitGroup
	isClosed     int32
	ch           chan *PingResult
	cached_bytes []byte
	newRequest   func(id, seqnum uint16, msglen int, filler, cached []byte) []byte
}

func newPinger(family int, network, laddr string, echo []byte, capacity int) (*Pinger, error) {
	c, err := net.ListenPacket(network, laddr)
	if err != nil {
		return nil, fmt.Errorf("ListenPacket(%q, %q) failed: %v", network, laddr, err)
	}

	if nil == echo || 0 == len(echo) {
		echo = []byte("gogogogo")
	}

	newRequest := newICMPv4EchoRequest
	if family == syscall.AF_INET6 {
		newRequest = newICMPv6EchoRequest
	}

	icmp := &Pinger{family: family,
		network:      network,
		seqnum:       6145,
		id:           uint16(os.Getpid() & 0xffff),
		echo:         echo,
		conn:         c,
		ch:           make(chan *PingResult, capacity),
		cached_bytes: make([]byte, 1024),
		newRequest:   newRequest}
	//icmp.Send("127.0.0.1", nil)
	go icmp.serve()
	icmp.wait.Add(1)
	return icmp, nil
}

func NewPinger(netwwork, laddr string, echo []byte, capacity int) (*Pinger, error) {
	if netwwork == "ip4:icmp" {
		return newPinger(syscall.AF_INET, netwwork, laddr, echo, capacity)
	}
	if netwwork == "ip6:ipv6-icmp" {
		return newPinger(syscall.AF_INET6, netwwork, laddr, echo, capacity)
	}
	return nil, errors.New("Unsupported network - " + netwwork)
}

func (self *Pinger) Close() error {
	if atomic.CompareAndSwapInt32(&self.isClosed, 0, 1) {
		self.conn.Close()
		self.wait.Wait()
		close(self.ch)
	}
	return nil
}

func (self *Pinger) GetChannel() <-chan *PingResult {
	return self.ch
}

func (self *Pinger) Send(raddr string, echo []byte) error {
	ra := net.ParseIP(raddr)
	if ra == nil {
		return fmt.Errorf("ParseIP failed: %v", raddr)
	}
	return self.SendWithIPAddr(&net.IPAddr{IP: ra}, self.id, self.seqnum, echo)
}

func (self *Pinger) SendWithIP(addr net.IP, echo []byte) error {
	return self.SendWithIPAddr(&net.IPAddr{IP: addr}, self.id, self.seqnum, echo)
}

func (self *Pinger) SendWithIPAddr(ra *net.IPAddr, id, seqnum uint16, echo []byte) error {
	self.seqnum++
	filler := echo
	if nil == filler {
		filler = self.echo
	}

	msglen := len(filler) + 8
	if msglen > 1024 {
		msglen = 2039
	}

	bytes := self.newRequest(id, seqnum, msglen, filler, self.cached_bytes)
	_, err := self.conn.WriteTo(bytes, ra)
	if err != nil {
		return fmt.Errorf("WriteTo failed: %v", err)
	}
	return nil
}

func (self *Pinger) Recv(timeout time.Duration) (net.Addr, []byte, error) {
	timer := time.NewTimer(timeout)
	select {
	case res := <-self.ch:
		timer.Stop()
		addr := res.Addr
		bs := res.Bytes
		err := res.Err

		res.Id = 0
		res.SeqNum = 0
		res.Addr = nil
		res.Bytes = nil
		res.Err = nil
		resultPool.Put(res)
		return addr, bs, err
	case <-timer.C:
		return nil, nil, errors.ErrTimeout
	}
	return nil, nil, errors.ErrTimeout
}

func (self *Pinger) serve() {
	defer self.wait.Done()

	var cached = resultArrayPool.Get().([]*PingResult)
	var sending int32 = 0
	for nil != self.conn {
		reply := make([]byte, 2048)
		l, ra, err := self.conn.ReadFrom(reply)
		if err != nil {
			self.ch <- &PingResult{Err: fmt.Errorf("ReadFrom failed: %v", err)}
			break
		}

		switch self.family {
		case syscall.AF_INET:
			if reply[0] != ICMP4_ECHO_REPLY {
				continue
			}
		case syscall.AF_INET6:
			if reply[0] != ICMP6_ECHO_REPLY {
				continue
			}
		}

		_, _, id, seqnum, bs := parseICMPEchoReply(reply[:l])
		pingResult := resultPool.Get().(*PingResult)
		pingResult.Addr = ra
		pingResult.Id = id
		pingResult.SeqNum = seqnum
		pingResult.Bytes = bs
		pingResult.Timestamp = time.Now()

		cached = append(cached, pingResult)

		if atomic.CompareAndSwapInt32(&sending, 0, 1) {
			self.wait.Add(1)
			go func(all []*PingResult) {
				defer func() {
					atomic.StoreInt32(&sending, 0)
					self.wait.Done()
				}()

				for _, pr := range all {
					self.ch <- pr
				}

				if len(all) > 10 {
					logger.Println("队列有点堵了 - ", len(all))
				}
				resultArrayPool.Put(all[:0])
			}(cached)

			cached = resultArrayPool.Get().([]*PingResult)
		}
	}
}

func newICMPEchoRequest(family int, id, seqnum uint16, msglen int, filler, cached []byte) []byte {
	if family == syscall.AF_INET6 {
		return newICMPv6EchoRequest(id, seqnum, msglen, filler, cached)
	}
	return newICMPv4EchoRequest(id, seqnum, msglen, filler, cached)
}

func newICMPv4EchoRequest(id, seqnum uint16, msglen int, filler, cached []byte) []byte {
	b := newICMPInfoMessage(id, seqnum, msglen, filler, cached)
	b[0] = ICMP4_ECHO_REQUEST

	// calculate Pinger checksum
	cklen := len(b)
	s := uint32(0)
	for i := 0; i < cklen-1; i += 2 {
		s += uint32(b[i+1])<<8 | uint32(b[i])
	}
	if cklen&1 == 1 {
		s += uint32(b[cklen-1])
	}
	s = (s >> 16) + (s & 0xffff)
	s = s + (s >> 16)
	// place checksum back in header; using ^= avoids the
	// assumption the checksum bytes are zero
	b[2] ^= uint8(^s & 0xff)
	b[3] ^= uint8(^s >> 8)

	return b
}

func newICMPv6EchoRequest(id, seqnum uint16, msglen int, filler, cached []byte) []byte {
	b := newICMPInfoMessage(id, seqnum, msglen, filler, cached)
	b[0] = ICMP6_ECHO_REQUEST
	return b
}

func newICMPInfoMessage(id, seqnum uint16, msglen int, filler, cached []byte) []byte {
	var b []byte
	if len(cached) > msglen {
		b = cached[0:msglen]
	} else {
		b = make([]byte, msglen)
	}
	copy(b[8:], bytes.Repeat(filler, (msglen-8)/len(filler)+1))
	b[0] = 0                    // type
	b[1] = 0                    // code
	b[2] = 0                    // checksum
	b[3] = 0                    // checksum
	b[4] = uint8(id >> 8)       // identifier
	b[5] = uint8(id & 0xff)     // identifier
	b[6] = uint8(seqnum >> 8)   // sequence number
	b[7] = uint8(seqnum & 0xff) // sequence number
	return b
}

func parseICMPEchoReply(b []byte) (t, code, id, seqnum uint16, body []byte) {
	t = uint16(b[0])
	code = uint16(b[1])
	id = uint16(b[4])<<8 | uint16(b[5])
	seqnum = uint16(b[6])<<8 | uint16(b[7])
	return t, code, id, seqnum, b[8:]
}

type Pingers struct {
	*Pinger
}

func NewPingers(echo []byte, capacity int) (*Pingers, error) {
	v4, e := NewPinger("ip4:icmp", "", echo, capacity)
	if nil != e {
		return nil, e
	}
	return &Pingers{v4}, nil
}
