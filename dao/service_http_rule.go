package dao

import (
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
	"github.com/liuyj/gateway/public"
)

type HttpRule struct {
	ID             int64  `json:"id" gorm:"primary_key"`
	ServiceID      int64  `json:"service_id" gorm:"column:service_id" description:"服务id"`
	RuleType       int    `json:"rule_type" gorm:"column:rule_type" description:"匹配类型 domain=域名, url_prefix=url前缀"`
	Rule           string `json:"rule" gorm:"column:rule" description:"type=domain表示域名，type=url_prefix时表示url前缀"`
	NeedHttps      int    `json:"need_https" gorm:"column:need_https" description:"type=支持https 1=支持"`
	NeedWebsocket  int    `json:"need_websocket" gorm:"column:need_websocket" description:"启用websocket 1=启用"`
	NeedStripUri   int    `json:"need_strip_uri" gorm:"column:need_strip_uri" description:"启用strip_uri 1=启用"`
	UrlRewrite     string `json:"url_rewrite" gorm:"column:url_rewrite" description:"url重写功能，每行一个	"`
	HeaderTransfor string `json:"header_transfor" gorm:"column:header_transfor" description:"header转换支持增加(add)、删除(del)、修改(edit) 格式: add headname headvalue	"`
}

func (hr *HttpRule) TableName() string {
	return "gateway_service_http_rule"
}

func (hr *HttpRule) Find(c *gin.Context, db *gorm.DB, search *HttpRule) (*HttpRule, error) {
	result := &HttpRule{}
	err := db.SetCtx(public.GetGinTraceContext(c)).Where(search).Find(result).Error
	return result, err
}

func (hr *HttpRule) Save(c *gin.Context, db *gorm.DB) error {
	if err := db.SetCtx(public.GetGinTraceContext(c)).Save(hr).Error; err != nil {
		return err
	}
	return nil
}
