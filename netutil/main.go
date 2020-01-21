package netutil

import (
	"encoding/binary"
	"errors"
	"net"
	"strings"
)

type IPChecker interface {
	Contains(net.IP) bool
}

type ipRangeChecker struct {
	start, end uint32
}

func (r *ipRangeChecker) String() string {
	var a, b [4]byte
	binary.BigEndian.PutUint32(a[:], r.start)
	binary.BigEndian.PutUint32(b[:], r.end)
	return net.IP(a[:]).String() + "-" +
		net.IP(b[:]).String()
}

func (r *ipRangeChecker) Contains(ip net.IP) bool {
	if ip.To4() == nil {
		return false
	}

	v := binary.BigEndian.Uint32(ip.To4())
	return r.start <= v && v <= r.end
}

func IPRangeChecker(start, end net.IP) (IPChecker, error) {
	if start.To4() == nil {
		return nil, errors.New("ip range 不支持 IPv6")
	}
	if end.To4() == nil {
		return nil, errors.New("ip range 不支持 IPv6")
	}
	s := binary.BigEndian.Uint32(start.To4())
	e := binary.BigEndian.Uint32(end.To4())
	return &ipRangeChecker{start: s, end: e}, nil
}

func IPRangeCheckerWith(start, end string) (IPChecker, error) {
	s := net.ParseIP(start)
	if s == nil {
		return nil, errors.New(start + " is invalid address")
	}
	e := net.ParseIP(end)
	if e == nil {
		return nil, errors.New(end + " is invalid address")
	}
	return IPRangeChecker(s, e)
}

var (
	_ IPChecker = &net.IPNet{}
	_ IPChecker = &ipRangeChecker{}

	ErrInvalidIPRange = errors.New("invalid ip range")
)

func ToChecker(s string) (IPChecker, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	if strings.Contains(s, "-") {
		ss := strings.Split(s, "-")
		if len(ss) != 2 {
			return nil, ErrInvalidIPRange
		}
		checker, err := IPRangeCheckerWith(ss[0], ss[1])
		if err != nil {
			return nil, ErrInvalidIPRange
		}
		return checker, nil
	}

	if strings.Contains(s, "/") {
		_, cidr, err := net.ParseCIDR(s)
		if err != nil {
			return nil, ErrInvalidIPRange
		}
		return cidr, nil
	}

	checker, err := IPRangeCheckerWith(s, s)
	if err != nil {
		return nil, ErrInvalidIPRange
	}
	return checker, nil
}

func ToCheckers(ipList []string) ([]IPChecker, error) {
	var ingressIPList []IPChecker
	for _, s := range ipList {
		checker, err := ToChecker(s)
		if err != nil {
			return nil, err
		}
		if checker != nil {
			ingressIPList = append(ingressIPList, checker)
		}
	}
	return ingressIPList, nil
}
