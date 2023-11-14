package controller

import (
	"FGateWay/dao"
	"FGateWay/dto"
	"FGateWay/golang_common/lib"
	"FGateWay/middleware"
	"FGateWay/public"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AdminController struct{}

func AdminRegister(group *gin.RouterGroup) {
	adminInfo := &AdminController{}
	group.GET("/admin_info", adminInfo.AdminInfo)
	group.POST("/change_pwd", adminInfo.ChangePwd)
}

// AdminInfo Admin godoc
// @Summary 管理员信息
// @Description 管理员接口
// @Tags 管理员接口
// @ID /admin/admin_info
// @Accept json
// @Produce json
// @Success 200 {object} middleware.Response{data=dto.AdminInfoOutput{}}  "success"
// @Router /admin/admin_info [get]
func (administer *AdminController) AdminInfo(c *gin.Context) {
	//1.读取sessionKey对应的JSON 转换为结构体
	//2.取出数据然后封装输出结构体

	sess := sessions.Default(c)
	sessInfo := sess.Get(public.AdminSessionInfoKey)
	//1.读取session对应的json,转换为结构体
	//2.取出数据然后封装输出结构体
	adminSessionInfo := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(fmt.Sprint(sessInfo)), &adminSessionInfo); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	out := &dto.AdminInfoOutput{
		ID:           adminSessionInfo.ID,
		UserName:     adminSessionInfo.UserName,
		LogInTime:    adminSessionInfo.LogInTime,
		Avatar:       "",
		Introduction: "",
		Roles:        nil,
	}
	middleware.ResponseSuccess(c, out)
}

// ChangePwd godoc
// @Summary 改变密码
// @Description 管理员接口
// @ID /admin/change_pwd
// @Tags 管理员接口
// @Accept json
// @Produce json
// @Param body body dto.ChangePwdInput true "body"
// @Success 200 {object} middleware.Response{data=string}  "success"
// @Router /admin/change_pwd [post]
func (administer *AdminController) ChangePwd(c *gin.Context) {
	params := &dto.ChangePwdInput{}
	if err := params.BindValidParm(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}
	//sessinfo读取用户信息到结构体 sessinfo
	sess := sessions.Default(c)
	sessInfo := sess.Get(public.AdminSessionInfoKey)
	//1.读取session对应的json,转换为结构体
	//2.取出数据然后封装输出结构体

	adminSessionInfo := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(fmt.Sprint(sessInfo)), &adminSessionInfo); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	//sessinfo.id 读取数据库信息 admininfo
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	adminfo := &dao.Admin{}
	adminfo, err = adminfo.Find(c, tx, &dao.Admin{
		UserName: adminSessionInfo.UserName,
	})
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	//saltPassword
	saltPassword := public.GenSaltPassword(adminfo.Salt, params.Password)
	adminfo.Password = saltPassword
	//执行数据库保存
	if err = adminfo.Save(c, tx); err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, "")
}
