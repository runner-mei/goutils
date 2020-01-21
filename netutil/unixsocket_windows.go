package netutil

import (
	"context"
	"net"
	"strings"
	"time"

	winio "github.com/Microsoft/go-winio"
)

func NewUnixListener(network, file string) (net.Listener, error) {
	if !strings.HasPrefix(file, `\\`) {
		file = MakePipenameWindows(file)
	}

	listener, err := winio.ListenPipe(file, nil)
	if err != nil {
		return nil, err
	}
	return listener, nil
}

func Dial(network, addr string) (net.Conn, error) {
	if IsUnixsocket(network) {
		return winio.DialPipe(addr, nil)
	}

	return net.Dial(network, addr)
}

func DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	if IsUnixsocket(network) {
		return winio.DialPipeContext(ctx, addr)
	}

	var d net.Dialer
	return d.DialContext(ctx, network, addr)
}

func DialTimeout(network, addr string, timeout time.Duration) (net.Conn, error) {
	if IsUnixsocket(network) {
		return winio.DialPipe(addr, &timeout)
	}

	return net.DialTimeout(network, addr, timeout)
}

// Dial connects to the address on the named network.
func (d *Dialer) Dial(network, address string) (net.Conn, error) {
	if IsUnixsocket(network) {
		return winio.DialPipe(address, nil)
	}

	return d.Dialer.DialContext(context.Background(), network, address)
}

// DialContext connects to the address on the named network using
// the provided context.
func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if IsUnixsocket(network) {
		return winio.DialPipeContext(ctx, address)
	}
	return d.Dialer.DialContext(ctx, network, address)
}

func (dialer *HttpDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
		port = "80"
	}

	if host == UNIXSOCKET {
		pipeName := MakePipenameWindows(port)
		return winio.DialPipeContext(ctx, pipeName)
	}

	if dialer.DialWithContext == nil {
		var dialer net.Dialer
		return dialer.DialContext(ctx, network, addr)
	}
	return dialer.DialWithContext(ctx, network, addr)
}
