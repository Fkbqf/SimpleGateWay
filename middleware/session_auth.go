package middleware

import (
	"FGateWay/public"
	"errors"
	"fmt"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func SessionAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		fmt.Println("session ", session.Get(public.AdminSessionInfoKey).(string))
		if adminInfo, ok := session.Get(public.AdminSessionInfoKey).(string); !ok || adminInfo == "" {

			ResponseError(c, InternalErrorCode, errors.New("user not login"))
			c.Abort()
			return
		}
		c.Next()
	}
}
