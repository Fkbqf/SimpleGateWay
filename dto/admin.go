package dto

import (
	"FGateWay/public"
	"github.com/gin-gonic/gin"
	"time"
)

type AdminInfoOutput struct {
	ID           int       `json:"id" `
	UserName     string    `json:"user_name" `
	LogInTime    time.Time `json:"login_time"`
	Avatar       string    `json:"avatar"`
	Introduction string    `json:"introduction"`
	Roles        []string  `json]:"roles"`
}

type ChangePwdInput struct {
	Password string `json:"password" form:"password" comment:"密码" example:"123456" validate:"required"`
}

func (param *ChangePwdInput) BindValidParm(c *gin.Context) error {
	return public.DefaultGetValidParams(c, param)
}
