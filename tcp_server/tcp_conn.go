package tcp_server

import (
	"context"
	"fmt"
	"net"
	"runtime"
)

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (l tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := l.AcceptTCP()
	if err != nil {
		return nil, err
	}
	return tc, nil
}

type conn struct {
	server     *TcpServer
	cancelCtx  context.CancelCauseFunc
	rwc        net.Conn
	remoteAddr string
}

func (c *conn) serve(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil && err != ErrAbortHandler {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			fmt.Printf("tcp: panic serving %v: %v\n%s", c.remoteAddr, err, buf)
		}
		c.close()
	}()
	c.remoteAddr = c.rwc.RemoteAddr().String()

}

func (c *conn) close() {
	c.rwc.Close()
}
