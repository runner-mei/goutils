package netutil

import (
	"net"
	"os"
	"syscall"
	"testing"
	"time"
)

var icmpTests = []struct {
	network string
	laddr   string
	raddr   string
	ipv6    bool // test with underlying AF_INET6 socket
}{
	{"ip4:icmp", "", "127.0.0.1", false},
	//{"ip6:icmp", "", "::1", true},
}

func getfamily(network string) int {
	if network == "ip4:icmp" {
		return syscall.AF_INET
	}
	if network == "ip6:icmp" {
		return syscall.AF_INET6
	}
	return syscall.AF_UNSPEC
}

func TestICMP(t *testing.T) {
	//if os.Getuid() != 0 {
	//	t.Logf("test disabled; must be root")
	//	return
	//}

	seqnum := uint16(6145)
	for _, tt := range icmpTests {
		id := uint16(os.Getpid() & 0xffff)
		seqnum++

		echo := newICMPEchoRequest(getfamily(tt.network), id, seqnum, 128, []byte("Go Go Gadget Ping!!!"), nil)
		exchangeICMPEcho(t, tt.network, tt.laddr, tt.raddr, echo)
	}
}

func exchangeICMPEcho(t *testing.T, network, laddr, raddr string, echo []byte) {
	c, err := net.ListenPacket(network, laddr)
	if err != nil {
		t.Errorf("ListenPacket(%q, %q) failed: %v", network, laddr, err)
		return
	}
	c.SetDeadline(time.Now().Add(100 * time.Millisecond))
	defer c.Close()

	ra, err := net.ResolveIPAddr(network, raddr)
	if err != nil {
		rip := net.ParseIP(raddr)
		if rip == nil {
			t.Errorf("ResolveIPAddr(%q, %q) failed: %v", network, raddr, err)
			return
		}
		ra = &net.IPAddr{IP: rip}
	}

	//waitForReady := make(chan bool)
	//go icmpEchoTransponder(t, network, raddr, waitForReady)
	//<-waitForReady

	_, err = c.WriteTo(echo, ra)
	if err != nil {
		t.Errorf("WriteTo failed: %v", err)
		return
	}

	reply := make([]byte, 256)
	for {
		_, _, err := c.ReadFrom(reply)
		if err != nil {
			t.Errorf("ReadFrom failed: %v", err)
			return
		}
		switch family(ra) {
		case syscall.AF_INET:
			if reply[0] != ICMP4_ECHO_REPLY {
				continue
			}
		case syscall.AF_INET6:
			if reply[0] != ICMP6_ECHO_REPLY {
				continue
			}
		}
		_, _, xid, xseqnum, _ := parseICMPEchoReply(echo)
		_, _, rid, rseqnum, _ := parseICMPEchoReply(reply)
		if rid != xid || rseqnum != xseqnum {
			t.Errorf("ID = %v, Seqnum = %v, want ID = %v, Seqnum = %v", rid, rseqnum, xid, xseqnum)
			return
		}
		break
	}
}

// func icmpEchoTransponder(t *testing.T, network, raddr string, waitForReady chan bool) {
// 	c, err := net.Dial(network, raddr)
// 	if err != nil {
// 		waitForReady <- true
// 		t.Errorf("Dial(%q, %q) failed: %v", network, raddr, err)
// 		return
// 	}
// 	c.SetDeadline(time.Now().Add(100 * time.Millisecond))
// 	defer c.Close()
// 	waitForReady <- true

// 	echo := make([]byte, 256)
// 	var nr int
// 	for {
// 		nr, err = c.Read(echo)
// 		if err != nil {
// 			t.Errorf("Read failed: %v", err)
// 			return
// 		}
// 		switch family(nil) {
// 		case syscall.AF_INET:
// 			if echo[0] != ICMP4_ECHO_REQUEST {
// 				continue
// 			}
// 		case syscall.AF_INET6:
// 			if echo[0] != ICMP6_ECHO_REQUEST {
// 				continue
// 			}
// 		}
// 		break
// 	}

// 	switch family(c.RemoteAddr()) {
// 	case syscall.AF_INET:
// 		echo[0] = ICMP4_ECHO_REPLY
// 	case syscall.AF_INET6:
// 		echo[0] = ICMP6_ECHO_REPLY
// 	}

// 	_, err = c.Write(echo[:nr])
// 	if err != nil {
// 		t.Errorf("Write failed: %v", err)
// 		return
// 	}
// }
