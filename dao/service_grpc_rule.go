package dao

import (
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
	"github.com/liuyj/gateway/public"
)

type GrpcRule struct {
	ID             int64  `json:"id" gorm:"primary_key"`
	ServiceID      int64  `json:"service_id" gorm:"column:service_id" description:"服务id	"`
	Port           int    `json:"port" gorm:"column:port" description:"端口	"`
	HeaderTransfor string `json:"header_transfor" gorm:"column:header_transfor" description:"header转换支持增加(add)、删除(del)、修改(edit) 格式: add headname headvalue"`
}

func (t *GrpcRule) TableName() string {
	return "gateway_service_grpc_rule"
}

func (*GrpcRule) Find(c *gin.Context, db *gorm.DB, search *GrpcRule) (*GrpcRule, error) {
	result := &GrpcRule{}
	err := db.SetCtx(public.GetGinTraceContext(c)).Where(search).Find(result).Error
	return result, err
}

func (gr *GrpcRule) Save(c *gin.Context, db *gorm.DB) error {
	if err := db.SetCtx(public.GetGinTraceContext(c)).Save(gr).Error; err != nil {
		return err
	}
	return nil
}
