package dto

import (
	"FGateWay/public"
	"github.com/gin-gonic/gin"
	"time"
)

type AdminSessionInfo struct {
	ID        int       `json:"id" `
	UserName  string    `json:"user_name"`
	LogInTime time.Time `json:"login_time"`
}

type AdminLoginInput struct {
	UserName string `json:"username" form:"username" comment:"姓名" example:"姓名" validate:"required,valid_username"`
	Password string `json:"password" form:"password" comment:"密码" example:"admin" validate:"required"`
}

func (param *AdminLoginInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, param)
}

type AdminLoginOput struct {
	Token string `json:"token" form:"token" comment:"token" example:"token" validate:""`
}
