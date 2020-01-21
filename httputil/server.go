package httputil

import (
	"net"
	"net/http"
	"time"

	"github.com/runner-mei/goutils/netutil"
)

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type TcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln TcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

func RunTLS(network, addr, certFile, keyFile string, engine http.Handler) (err error) {
	if network == "" {
		network = "tcp"
	}

	if !netutil.IsUnixsocket(network) {
		return http.ListenAndServeTLS(addr, certFile, keyFile, engine)
	}

	listener, err := netutil.NewUnixListener(network, addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	return http.ServeTLS(listener, engine, certFile, keyFile)
}

func RunHTTP(network, addr string, engine http.Handler) (err error) {
	if network == "" {
		network = "tcp"
	}

	if !netutil.IsUnixsocket(network) {
		return http.ListenAndServe(addr, engine)
	}

	listener, err := netutil.NewUnixListener(network, addr)
	if err != nil {
		return err
	}
	defer listener.Close()
	return http.Serve(listener, engine)
}

func RunServer(srv *http.Server, network, addr string) error {
	ln, err := netutil.Listen(network, addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	tcpListener, ok := ln.(*net.TCPListener)
	if ok {
		return srv.Serve(TcpKeepAliveListener{tcpListener})
	}
	return srv.Serve(ln)
}

func RunServerTLS(srv *http.Server, network, addr, certFile, keyFile string) error {
	ln, err := netutil.Listen(network, addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	tcpListener, ok := ln.(*net.TCPListener)
	if ok {
		return srv.ServeTLS(TcpKeepAliveListener{tcpListener}, certFile, keyFile)
	}

	return srv.ServeTLS(ln, certFile, keyFile)
}
