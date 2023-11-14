package http_proxy_middleware

import (
	"FGateWay/dao"
	"FGateWay/middleware"
	"github.com/gin-gonic/gin"
)

func HttpAccessModeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		service, err := dao.ServiceManagerHandler.HTTPAccessMode(c)
		if err != nil {
			middleware.ResponseError(c, 1001, err)
			return
		}
		c.Set("service", service)
		c.Next()
	}
}
