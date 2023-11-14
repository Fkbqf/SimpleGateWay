package controller

import (
	"FGateWay/dao"
	"FGateWay/dto"
	"FGateWay/golang_common/lib"
	"FGateWay/middleware"
	"FGateWay/public"
	"encoding/json"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"time"
)

type AdminLoginController struct{}

func AdminLoginControllerRegister(group *gin.RouterGroup) {
	adminLogin := &AdminLoginController{}
	group.POST("/login", adminLogin.AdminLogin)
	group.GET("/logout", adminLogin.AdminLoginOut)
}

// AdminLogin godoc
// @Summary 管理员登陆
// @Description 管理员接口
// @ID /admin_login/login
// @Tags 管理员接口
// @Accept json
// @Produce json
// @Param body body dto.AdminLoginInput true "body"
// @Success 200 {object} middleware.Response{data=dto.AdminLoginOput}  "success"
// @Router /admin_login/login [post]
func (adminlogin *AdminLoginController) AdminLogin(c *gin.Context) {
	params := &dto.AdminLoginInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	//1.parms.Username 取得管理员信息 admininfo
	//2.adminfo.salt + params.password sha256 =》saltpassword
	//3.saltpassword==adminfo.password
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	admin := &dao.Admin{}
	admin, err = admin.LoginCheck(c, tx, params)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	sessInfo := &dto.AdminSessionInfo{
		ID:        admin.Id,
		UserName:  admin.UserName,
		LogInTime: time.Now(),
	}

	//设置session
	sessBites, err := json.Marshal(sessInfo)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}

	sess := sessions.Default(c)
	sess.Set(public.AdminSessionInfoKey, string(sessBites))
	sess.Save()
	out := &dto.AdminLoginOput{Token: admin.UserName}
	middleware.ResponseSuccess(c, out)
}

// AdminLogin godoc
// @Summary 管理员退出
// @Description 管理员接口
// @Tags 管理员接口
// @ID /admin_login/logout
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin_login/logout [get]
func (adminlogin *AdminLoginController) AdminLoginOut(c *gin.Context) {
	sess := sessions.Default(c)
	sess.Delete(public.AdminSessionInfoKey)
	sess.Save()
	middleware.ResponseSuccess(c, "")
}
