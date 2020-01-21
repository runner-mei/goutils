package netutil

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"time"
)

func NewUnixListener(network, file string) (net.Listener, error) {
	if !filepath.IsAbs(file) {
		file = MakePipenameUnix(file)
	}

	if err := os.Remove(file); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	listener, err := net.Listen(network, file)
	if err != nil {
		return nil, err
	}
	os.Chmod(file, 0777)
	return listener, nil
}

func Dial(network, addr string) (net.Conn, error) {
	return net.Dial(network, addr)
}

func DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	var d net.Dialer
	return d.DialContext(ctx, network, addr)
}

func DialTimeout(network, addr string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

// Dial connects to the address on the named network.
func (d *Dialer) Dial(network, address string) (net.Conn, error) {
	return d.Dialer.DialContext(context.Background(), network, address)
}

// DialContext connects to the address on the named network using
// the provided context.
func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d.Dialer.DialContext(ctx, network, address)
}

func (dialer *HttpDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
		port = "80"
	}

	if host == UNIXSOCKET {
		network = "unix"
		addr = MakePipenameUnix(port)
	}

	if dialer.DialWithContext == nil {
		var dialer net.Dialer
		return dialer.DialContext(ctx, network, addr)
	}
	return dialer.DialWithContext(ctx, network, addr)
}
