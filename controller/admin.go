package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/e421083458/golang_common/lib"
	"github.com/liuyj/gateway/dao"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/liuyj/gateway/dto"
	"github.com/liuyj/gateway/middleware"
	"github.com/liuyj/gateway/public"
)

type AdminInfoController struct{}

func AdminRegister(group *gin.RouterGroup) {
	adminInfo := &AdminInfoController{}
	group.GET("/admin_info", adminInfo.AdminInfo)
	group.POST("/change_pwd", adminInfo.ChangePwd)
}

// ChangePwd godoc
// @Summary 管理员密码修改
// @Description 管理员密码修改
// @Tags 管理员修改密码接口
// @ID /admin/change_pwd
// @Accept  json
// @Produce  json
// @Param body body dto.ChangePwdInput true "body"
// @Success 200 "success"
// @Router /admin/change_pwd [post]
func (*AdminInfoController) ChangePwd(c *gin.Context) {
	params := &dto.ChangePwdInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, http.StatusBadRequest, err)
		return
	}
	//1. get admin info from session
	session := sessions.Default(c)
	sessionInfo := session.Get(public.AdminSessionInfoKey)
	adminSessionInfo := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(fmt.Sprint(sessionInfo)), adminSessionInfo); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	admin := &dao.Admin{}
	admin, err = admin.Find(c, db, &dao.Admin{UserName: adminSessionInfo.UserName})
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	//2. generate new password
	newPassword := public.GenerateSaltPassword(admin.Salt, params.Password)
	admin.Password = newPassword

	//3. save new password into the db
	err = admin.Save(c, db)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, "change successfully")
}

// AdminInfo godoc
// @Summary 管理员信息
// @Description 管理员信息
// @Tags 管理员信息接口
// @ID /admin/admin_info
// @Accept  json
// @Produce  json
// @Success 200 "success"
// @Router /admin/admin_info [get]
func (*AdminInfoController) AdminInfo(c *gin.Context) {
	session := sessions.Default(c)
	sessionInfo := session.Get(public.AdminSessionInfoKey)

	if sessionInfo == nil {
		log.Printf("the sessionInfo is empty")
	} else {
		log.Printf("the sessionInfo is not empty")
		log.Printf("%v", sessionInfo)
	}

	adminSessionInfo := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(fmt.Sprint(sessionInfo)), adminSessionInfo); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}
	//convert adminSessionInfo into output form
	out := &dto.AdminInfoOutput{
		ID:           adminSessionInfo.ID,
		Name:         adminSessionInfo.UserName,
		LoginTime:    adminSessionInfo.LoginTime,
		Avatar:       "https://images.cnblogs.com/cnblogs_com/tanshaoshenghao/1432314/o_2005240121381.gif",
		Introduction: "super admin is me",
		Roles:        []string{"admin"},
	}
	middleware.ResponseSuccess(c, out)
}
