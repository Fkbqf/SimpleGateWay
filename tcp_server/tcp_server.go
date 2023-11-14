package tcp_server

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrServerClosed     = errors.New("tcp: Server closed")
	ErrAbortHandler     = errors.New("tcp:abort TcpHandler")
	ServerContextKey    = &contextKey{"tcp-server"}
	LocalAddrContextKey = &contextKey{"local-addr"}
)

type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "tcp_proxy context value " + k.name
}

type onceCloseLister struct {
	net.Listener
	once     sync.Once
	closeErr error
}

func (oc *onceCloseLister) Close() error {
	oc.once.Do(oc.close)
	return oc.closeErr
}

func (oc *onceCloseLister) close() {
	oc.closeErr = oc.Listener.Close()
}

type TCPHandler interface {
	ServerTcp(ctx context.Context, conn net.Conn)
}

type TcpServer struct {
	Addr    string
	Handler TCPHandler
	err     error
	BaseCtx context.Context

	WriteTimeOut     time.Duration
	ReadTimeout      time.Duration
	KeepAliveTimeout time.Duration

	mu         sync.Mutex
	inShutdown int32
	doneChan   chan struct{}
	l          *onceCloseLister
}

func (srv *TcpServer) ListenAndServer() error {
	if srv.shuttingDown() {
		return ErrServerClosed
	}
	if srv.doneChan == nil {
		srv.doneChan = make(chan struct{})
	}
	addr := srv.Addr
	if addr == "" {
		return errors.New("need addr")
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return srv.Serve(tcpKeepAliveListener{
		ln.(*net.TCPListener),
	})
}

func (srv *TcpServer) Close() error {
	atomic.StoreInt32(&srv.inShutdown, 1)
	close(srv.doneChan) //关闭channel
	srv.l.close()
	return nil
}

func (srv *TcpServer) shuttingDown() bool {
	return atomic.LoadInt32(&srv.inShutdown) != 0
}

func (srv *TcpServer) newConn(rwc net.Conn) *conn {
	c := &conn{
		server: srv,
		rwc:    rwc,
	}
	//设置参数
	if d := c.server.ReadTimeout; d != 0 {
		c.rwc.SetReadDeadline(time.Now().Add(d))
	}
	if d := c.server.WriteTimeOut; d != 0 {
		c.rwc.SetWriteDeadline(time.Now().Add(d))
	}

	if d := c.server.KeepAliveTimeout; d != 0 {
		if tcpConn, ok := c.rwc.(*net.TCPConn); ok {
			tcpConn.SetKeepAlive(true)
			tcpConn.SetKeepAlivePeriod(d)
		}
	}
	return c
}

func (srv *TcpServer) getDoneChan() <-chan struct{} {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	if srv.doneChan == nil {
		srv.doneChan = make(chan struct{})
	}
	return srv.doneChan
}

func (srv *TcpServer) Serve(l net.Listener) error {
	srv.l = &onceCloseLister{Listener: l}
	defer srv.l.Close() //执行listner关闭
	if srv.BaseCtx == nil {
		srv.BaseCtx = context.Background()
	}
	baseCtx := srv.BaseCtx

	ctx := context.WithValue(baseCtx, ServerContextKey, srv)
	for {
		rw, e := l.Accept()
		if e != nil {
			select {
			case <-srv.getDoneChan():
				return ErrServerClosed
			default:
			}
			fmt.Printf("accept fail, err: %v\n", e)
			continue
		}
		c := srv.newConn(rw)
		go c.serve(ctx)
	}
}

func ListenAndServe(addr string, handler TCPHandler) error {
	server := &TcpServer{Addr: addr, Handler: handler, doneChan: make(chan struct{})}
	return server.ListenAndServer()
}
