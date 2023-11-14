package tcp_proxy_router

import (
	"FGateWay/dao"
	"FGateWay/reverse_proxy"
	"FGateWay/tcp_proxy_middleware"
	"FGateWay/tcp_server"
	"context"
	"fmt"
	"log"
	"net"
)

var tcpServerList = []*tcp_server.TcpServer{}

type tcpHandler struct {
}

func (t *tcpHandler) ServerTcp(ctx context.Context, src net.Conn) {
	src.Write([]byte("tcpHandler\n"))
}

func TcpServerRun() {
	serviceList := dao.ServiceManagerHandler.GetTcpServiceList()
	for _, serviItem := range serviceList {
		tempItem := serviItem

		go func(serviceDetail *dao.ServiceDetail) {
			addr := fmt.Sprintf(":%d", serviceDetail.TCPRule.Port)
			rb, err := dao.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
			if err != nil {
				log.Fatalf(" [INFO] GetTcpLoadBalancer %v err:%v\n", addr, err)
				return
			}

			router := tcp_proxy_middleware.NewTcpSliceRouter()
			router.Group("/").Use(
				tcp_proxy_middleware.TCPFlowCountMiddleware(),
				tcp_proxy_middleware.TCPFlowLimitMiddleware(),
				tcp_proxy_middleware.TCPWhiteListMiddleware(),
				tcp_proxy_middleware.TCPBlackListMiddleware())

			routerhanler := tcp_proxy_middleware.NewTcpSliceRouterHandler(
				func(c *tcp_proxy_middleware.TcpSliceRouterContext) tcp_server.TCPHandler {
					return reverse_proxy.NewTcpLoadBalanceReverseProxy(c, rb)
				}, router)
			baseCtx := context.WithValue(context.Background(), "service", serviceDetail)
			tcpserver := &tcp_server.TcpServer{
				Addr:    addr,
				Handler: routerhanler,
				BaseCtx: baseCtx,
			}

			tcpServerList = append(tcpServerList, tcpserver)
			log.Printf(" [INFO] tcp_proxy_run %v\n", addr)
			if err := tcpserver.ListenAndServer(); err != nil && err != tcp_server.ErrServerClosed {
				log.Fatalf(" [INFO] tcp_proxy_run %v err:%v\n", addr, err)
			}
		}(tempItem)

	}
}
func TcpServerStop() {
	for _, tcpServer := range tcpServerList {
		tcpServer.Close()
		log.Printf(" [INFO] tcp_proxy_stop %v stopped\n", tcpServer.Addr)
	}
}
