package http_proxy_middleware

import (
	"FGateWay/dao"
	"FGateWay/middleware"
	"FGateWay/reverse_proxy"
	"errors"
	"github.com/gin-gonic/gin"
)

func HTTPReverseProxyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serviceInterface.(*dao.ServiceDetail)

		lb, err := dao.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
		if err != nil {
			middleware.ResponseError(c, 2002, err)
			c.Abort()
			return
		}

		trans, err := dao.TransportorHandler.GetTrans(serviceDetail)
		if err != nil {
			middleware.ResponseError(c, 2003, err)
			c.Abort()
			return
		}
		proxy := reverse_proxy.NewLoadBalanceReverseProxy(c, lb, trans)
		proxy.ServeHTTP(c.Writer, c.Request)
		c.Abort()
		return
	}

}
