package main

import (
	"flag"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/runner-mei/goutils/netutil"
)

var (
	msg     = flag.String("msg", "gogogogo", "the body of icmp message, default: 'gogogogo'")
	laddr   = flag.String("laddr", "", "the address of bind, default: ''")
	network = flag.String("network", "ip4:icmp", "the family of address, default: 'ip4:icmp'")
)

func main() {
	flag.Parse()

	targets := flag.Args()
	if nil == targets || 1 != len(targets) {
		flag.Usage()
		return
	}

	icmp, err := netutil.NewPinger(*network, *laddr, []byte(*msg), 256)
	if nil != err {
		fmt.Println(err)
		return
	}
	defer func() {
		fmt.Println("exit")
		icmp.Close()
	}()

	ip_range, err := netutil.ParseIPRange(targets[0])
	if nil != err {
		fmt.Println(err)
		return
	}

	var is_stopped int32
	go func() {
		for ip_range.HasNext() {
			err = icmp.Send(ip_range.Current().String(), nil)
			if nil != err {
				fmt.Println(err)
				break
			}

			time.Sleep(100 * time.Millisecond)
		}
		atomic.StoreInt32(&is_stopped, 1)
	}()

	for {
		ra, _, err := icmp.Recv(10 * time.Second)
		if nil != err {
			if !errors.IsTimeout(err) {
				fmt.Println(err)
			} else if 0 == atomic.LoadInt32(&is_stopped) {
				continue
			}
			return
		}
		fmt.Println(ra)
	}

}
