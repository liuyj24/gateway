package dao

import (
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
	"github.com/liuyj/gateway/public"
)

type AccessControl struct {
	ID                int64  `json:"id" gorm:"primary_key"`
	ServiceID         int64  `json:"service_id" gorm:"column:service_id" description:"服务id"`
	OpenAuth          int    `json:"open_auth" gorm:"column:open_auth" description:"是否开启权限 1=开启"`
	BlackList         string `json:"black_list" gorm:"column:black_list" description:"黑名单ip	"`
	WhiteList         string `json:"white_list" gorm:"column:white_list" description:"白名单ip	"`
	WhiteHostName     string `json:"white_host_name" gorm:"column:white_host_name" description:"白名单主机	"`
	ClientIPFlowLimit int    `json:"clientip_flow_limit" gorm:"column:clientip_flow_limit" description:"客户端ip限流	"`
	ServiceFlowLimit  int    `json:"service_flow_limit" gorm:"column:service_flow_limit" description:"服务端限流	"`
}

func (ac *AccessControl) TableName() string {
	return "gateway_service_access_control"
}

func (ac *AccessControl) Find(c *gin.Context, db *gorm.DB, search *AccessControl) (*AccessControl, error) {
	result := &AccessControl{}
	err := db.SetCtx(public.GetGinTraceContext(c)).Where(search).Find(result).Error
	return result, err
}

func (ac *AccessControl) Save(c *gin.Context, db *gorm.DB) error {
	if err := db.SetCtx(public.GetGinTraceContext(c)).Save(ac).Error; err != nil {
		return err
	}
	return nil
}
