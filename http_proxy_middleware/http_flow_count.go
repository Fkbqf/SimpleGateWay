package http_proxy_middleware

import (
	"FGateWay/dao"
	"FGateWay/middleware"
	"FGateWay/public"
	"errors"
	"github.com/gin-gonic/gin"
)

func HTTpFlowCountMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serverInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 2001, errors.New("service not found"))
			c.Abort()
			return
		}

		serviceDetail := serverInterface.(*dao.ServiceDetail)

		//统计项 1 全站 2 服务 3 租户
		totalCounter, err := public.FlowCounterHandler.GetCounter(public.FlowTotal)
		if err != nil {
			middleware.ResponseError(c, 4001, err)
			c.Abort()
			return
		}
		totalCounter.Increase()

		serviceCounter, err := public.FlowCounterHandler.GetCounter(public.FlowServicePrefix + serviceDetail.Info.ServiceName)
		if err != nil {
			middleware.ResponseError(c, 4001, err)
			c.Abort()
			return
		}

		serviceCounter.Increase()
		c.Next()
	}
}
