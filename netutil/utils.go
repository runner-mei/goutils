package netutil

import (
	"errors"
	"net"
)

func IsInvalidAddress(addr string) (bool, error) {
	if 0 == len(addr) {
		return true, errors.New("ip is empty string.")
	}
	return IsInvalid(net.ParseIP(addr))
}

func IsInvalid(ip net.IP) (bool, error) {
	if nil == ip {
		return true, errors.New("ip is style error.")
	}
	if ip.IsUnspecified() {
		return true, errors.New("ip is Unspecified.")
	}
	if ip.IsLoopback() {
		return true, errors.New("ip is Loopback.")
	}
	if ip.IsInterfaceLocalMulticast() {
		return true, errors.New("ip is InterfaceLocalMulticast.")
	}
	if ip.IsLinkLocalMulticast() {
		return true, errors.New("ip is LinkLocalMulticast.")
	}
	if ip.IsLinkLocalUnicast() {
		return true, errors.New("ip is LinkLocalUnicast.")
	}
	if ip.IsMulticast() {
		return true, errors.New("ip is Multicast.")
	}
	return false, nil
}

var (
	BadMACs = [][]string{[]string{"00:00:00:00:00:00"},
		[]string{"00:00:00:00:00:01"},
		[]string{"ff:ff:ff:ff:ff:ff"},
		[]string{"00:ab:ab:ab:ab:ab"},
		[]string{"12:34:56:78:9a:bc"},
		[]string{"00:00:5e:00:00:00", "00:00:5e:ff:ff:ff"},
		[]string{"00:00:00:ff:ff:ff"},
		[]string{"cc:cc:cc:cc:cc:cc"}}
)

func IsValidPhysical(address string) bool {
	addr, e := net.ParseMAC(address)
	if nil != e {
		return false
	}
	return IsValidPhysicalAddress(addr)
}
func IsValidPhysicalAddress(address net.HardwareAddr) bool {
	if nil == address {
		return false
	}
	if 1 == address[0]&1 {
		return false
	}

	// 1. 所有字节值是一样的
	var invalid = true
	for _, b := range []byte(address)[1:] {
		if address[0] != b {
			invalid = false
			break
		}
	}
	if invalid {
		return false
	}

	// 2. 最后一个字节不为0，其它全为0
	for _, b := range []byte(address)[:len(address)-1] {
		if address[0] != b {
			invalid = false
			break
		}
	}
	if invalid {
		return false
	}

	// 3.前后全为0， 后面全为 ff
	excepted := byte(0)
	for _, b := range []byte(address) {
		if excepted != b {
			if 0 == excepted && 0xff == b {
				excepted = 0xff
				continue
			}

			invalid = false
			break
		}
	}
	if invalid {
		return false
	}

	ma := address.String()
	if "" == ma {
		return false
	}
	// if strings.HasPrefix(ma, "01") {
	// 	return false
	// }
	// if strings.HasPrefix(ma, "C1") {
	// 	return false
	// }
	// if strings.HasPrefix(ma, "33") {
	// 	return false
	// }

	for _, bma := range BadMACs {
		if len(bma) == 1 {
			if bma[0] == ma {
				return false
			}
			continue
		}
		if ma >= bma[0] && ma <= bma[1] {
			return false
		}
	}
	return true
}
