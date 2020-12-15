package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/liuyj/gateway/public"

	"github.com/gin-gonic/contrib/sessions"

	"github.com/liuyj/gateway/dao"

	"github.com/e421083458/golang_common/lib"

	"github.com/gin-gonic/gin"
	"github.com/liuyj/gateway/dto"
	"github.com/liuyj/gateway/middleware"
)

type AdminLoginController struct{}

func AdminLoginResgister(group *gin.RouterGroup) {
	adminLogin := &AdminLoginController{}
	group.POST("/login", adminLogin.AdminLogin)
	group.GET("/logout", adminLogin.AdminLogout)
}

// AdminLogout godoc
// @Summary 管理员退出登陆
// @Description 管理员退出登陆
// @Tags 管理员退出登录接口
// @ID /admin_login/logout
// @Accept  json
// @Produce  json
// @Success 200 "success"
// @Router /admin_login/logout [get]
func (*AdminLoginController) AdminLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete(public.AdminSessionInfoKey)
	session.Save()
	middleware.ResponseSuccess(c, "logout successfully")
}

// AdminLogin godoc
// @Summary 管理员登陆
// @Description 管理员登陆
// @Tags 管理员登录接口
// @ID /admin_login/login
// @Accept  json
// @Produce  json
// @Param body body dto.AdminLoginInput true "body"
// @Success 200 {object} middleware.Response{data=dto.AdminLoginOutput} "success"
// @Router /admin_login/login [post]
func (*AdminLoginController) AdminLogin(c *gin.Context) {
	params := &dto.AdminLoginInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, http.StatusBadRequest, err)
		return
	}
	//1. get admin info from db
	//2. check the password
	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	admin := &dao.Admin{}
	admin, err = admin.LoginCheck(c, db, params)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	//set sessionInfo
	sessionInfo := &dto.AdminSessionInfo{
		ID:        admin.Id,
		UserName:  admin.UserName,
		LoginTime: time.Now(),
	}
	sessionByte, err := json.Marshal(sessionInfo)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	session := sessions.Default(c)
	session.Set(public.AdminSessionInfoKey, string(sessionByte))
	session.Save()

	out := &dto.AdminLoginOutput{Token: admin.UserName}
	middleware.ResponseSuccess(c, out)
}
