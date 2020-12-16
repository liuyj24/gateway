package dao

import (
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
	"github.com/liuyj/gateway/public"
)

type TcpRule struct {
	ID        int64 `json:"id" gorm:"primary_key"`
	ServiceID int64 `json:"service_id" gorm:"column:service_id" description:"服务id	"`
	Port      int   `json:"port" gorm:"column:port" description:"端口	"`
}

func (t *TcpRule) TableName() string {
	return "gateway_service_tcp_rule"
}

func (*TcpRule) Find(c *gin.Context, db *gorm.DB, search *TcpRule) (*TcpRule, error) {
	result := &TcpRule{}
	err := db.SetCtx(public.GetGinTraceContext(c)).Where(search).Find(result).Error
	return result, err
}
