package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/liuyj/gateway/public"

	"github.com/liuyj/gateway/dao"

	"github.com/e421083458/golang_common/lib"

	"github.com/gin-gonic/gin"
	"github.com/liuyj/gateway/dto"
	"github.com/liuyj/gateway/middleware"
)

type ServiceController struct{}

func ServiceRegister(group *gin.RouterGroup) {
	service := &ServiceController{}
	group.GET("/service_list", service.ServiceList)
	group.POST("/service_delete", service.ServiceDelete)
	group.POST("/service_add_http", service.ServiceAddHTTP)
	group.POST("/service_update_http", service.ServiceUpdateHTTP)
}

// ServiceList godoc
// @Summary 服务列表
// @Description 服务列表
// @Tags 服务管理
// @ID /service/service_list
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Param page_size query int true "每页个数"
// @Param page_no query int true "当前页数"
// @Success 200 {object} middleware.Response{data=dto.ServiceListOutput} "success"
// @Router /service/service_list [get]
func (service *ServiceController) ServiceList(c *gin.Context) {
	param := &dto.ServiceListInput{}
	if err := param.BindValidParam(c); err != nil {
		middleware.ResponseError(c, http.StatusBadRequest, err)
		return
	}
	//get data from db
	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	serviceInfo := &dao.ServiceInfo{}
	list, total, err := serviceInfo.PageList(c, db, param)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	//convert db data struct into web visual data struct
	outputList := []dto.ServiceListItemOutput{}
	for _, item := range list {
		detail, err := item.ServiceDetail(c, db, &item)
		if err != nil {
			middleware.ResponseError(c, 2003, err)
			return
		}
		//设置服务地址
		serviceAddr := "unknow"
		clusterIP := lib.GetStringConf("base.cluster.cluster_ip")
		clusterPort := lib.GetStringConf("base.cluster.cluster_port")
		clusterSSLPort := lib.GetStringConf("base.cluster.cluster_ssl_port")

		//http(https)
		if detail.Info.LoadType == public.LoadTypeHTTP &&
			detail.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL &&
			detail.HTTPRule.NeedHttps == 1 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIP, clusterSSLPort, detail.HTTPRule.Rule)
		}
		//http(not https)
		if detail.Info.LoadType == public.LoadTypeHTTP &&
			detail.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL &&
			detail.HTTPRule.NeedHttps == 0 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIP, clusterPort, detail.HTTPRule.Rule)
		}
		//http
		if detail.Info.LoadType == public.LoadTypeHTTP &&
			detail.HTTPRule.RuleType == public.HTTPRuleTypeDomain {
			serviceAddr = detail.HTTPRule.Rule
		}
		//tcp
		if detail.Info.LoadType == public.LoadTypeTCP {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIP, detail.TCPRule.Port)
		}
		//grpc
		if detail.Info.LoadType == public.LoadTypeGRPC {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIP, detail.GRPCRule.Port)
		}
		ipList := detail.LoadBalance.GetIPListByModel()

		//todo qps qpd

		outItem := dto.ServiceListItemOutput{
			ID:          item.ID,
			LoadType:    item.LoadType,
			ServiceName: item.ServiceName,
			ServiceDesc: item.ServiceDesc,
			ServiceAddr: serviceAddr,
			Qps:         0,
			Qpd:         0,
			TotalNode:   len(ipList),
		}
		outputList = append(outputList, outItem)
	}
	out := &dto.ServiceListOutput{
		Total: total,
		List:  outputList,
	}
	middleware.ResponseSuccess(c, out)
}

// ServiceDelete godoc
// @Summary 服务删除
// @Description 服务删除
// @Tags 服务管理
// @ID /service/service_delete
// @Accept  json
// @Produce  json
// @Param id query string true "服务ID"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_delete [post]
func (service *ServiceController) ServiceDelete(c *gin.Context) {
	param := &dto.ServiceDeleteInput{}
	if err := param.BindValidParam(c); err != nil {
		middleware.ResponseError(c, http.StatusBadRequest, err)
		return
	}
	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	serviceInfo := &dao.ServiceInfo{ID: param.ID}
	result, err := serviceInfo.Find(c, db, serviceInfo)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	result.IsDelete = 1
	if err = result.Save(c, db); err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, "delete successfully!")
}

// ServiceAddHTTP godoc
// @Summary 添加HTTP服务
// @Description 添加HTTP服务
// @Tags 服务管理
// @ID /service/service_add_http
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddHTTPInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_add_http [post]
func (service *ServiceController) ServiceAddHTTP(c *gin.Context) {
	params := &dto.ServiceAddHTTPInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2000, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	tx = tx.Begin()

	serviceInfo := &dao.ServiceInfo{ServiceName: params.ServiceName}
	if _, err = serviceInfo.Find(c, tx, serviceInfo); err == nil {
		tx.Rollback()
		middleware.ResponseError(c, 2002, errors.New("服务已存在"))
		return
	}

	httpUrl := &dao.HttpRule{RuleType: params.RuleType, Rule: params.Rule}
	if _, err := httpUrl.Find(c, tx, httpUrl); err == nil {
		tx.Rollback()
		middleware.ResponseError(c, 2003, errors.New("服务接入前缀或域名已存在"))
		return
	}

	serviceModel := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err := serviceModel.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2005, err)
		return
	}
	//执行Save之后，serviceModel里面保存有新创建的id了
	httpRule := &dao.HttpRule{
		ServiceID:      serviceModel.ID,
		RuleType:       params.RuleType,
		Rule:           params.Rule,
		NeedHttps:      params.NeedHttps,
		NeedStripUri:   params.NeedStripUri,
		NeedWebsocket:  params.NeedWebsocket,
		UrlRewrite:     params.UrlRewrite,
		HeaderTransfor: params.HeaderTransfor,
	}
	if err := httpRule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         serviceModel.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		ClientIPFlowLimit: params.ClientipFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}

	loadbalance := &dao.LoadBalance{
		ServiceID:              serviceModel.ID,
		RoundType:              params.RoundType,
		IpList:                 params.IpList,
		WeightList:             params.WeightList,
		UpstreamConnectTimeout: params.UpstreamConnectTimeout,
		UpstreamHeaderTimeout:  params.UpstreamHeaderTimeout,
		UpstreamIdleTimeout:    params.UpstreamIdleTimeout,
		UpstreamMaxIdle:        params.UpstreamMaxIdle,
	}
	if err := loadbalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2008, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(c, "add http service successfully")
}

// ServiceUpdateHTTP godoc
// @Summary 修改HTTP服务
// @Description 修改HTTP服务
// @Tags 服务管理
// @ID /service/service_update_http
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateHTTPInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_update_http [post]
func (service *ServiceController) ServiceUpdateHTTP(c *gin.Context) {
	params := &dto.ServiceUpdateHTTPInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, http.StatusBadRequest, err)
		return
	}
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2001, errors.New("IP列表与权重列表数量不一致"))
		return
	}
	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	db.Begin()

	serviceInfo := &dao.ServiceInfo{ServiceName: params.ServiceName}
	serviceInfo, err = serviceInfo.Find(c, db, serviceInfo)
	if err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2003, errors.New("服务不存在"))
		return
	}
	detail, err := serviceInfo.ServiceDetail(c, db, serviceInfo)
	if err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2004, errors.New("服务不存在"))
		return
	}

	info := detail.Info
	info.ServiceDesc = params.ServiceDesc
	if err := info.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2005, err)
		return
	}

	httpRule := detail.HTTPRule
	httpRule.NeedHttps = params.NeedHttps
	httpRule.NeedStripUri = params.NeedStripUri
	httpRule.NeedWebsocket = params.NeedWebsocket
	httpRule.UrlRewrite = params.UrlRewrite
	httpRule.HeaderTransfor = params.HeaderTransfor
	if err := httpRule.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}

	accessControl := detail.AccessControl
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.ClientIPFlowLimit = params.ClientipFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}

	loadbalance := detail.LoadBalance
	loadbalance.RoundType = params.RoundType
	loadbalance.IpList = params.IpList
	loadbalance.WeightList = params.WeightList
	loadbalance.UpstreamConnectTimeout = params.UpstreamConnectTimeout
	loadbalance.UpstreamHeaderTimeout = params.UpstreamHeaderTimeout
	loadbalance.UpstreamIdleTimeout = params.UpstreamIdleTimeout
	loadbalance.UpstreamMaxIdle = params.UpstreamMaxIdle
	if err := loadbalance.Save(c, db); err != nil {
		db.Rollback()
		middleware.ResponseError(c, 2008, err)
		return
	}
	db.Commit()
	middleware.ResponseSuccess(c, "update successfully!")
}

func Test() {
	var a interface{}
	s, ok := a.(string)
	if ok {
		fmt.Println(s)
	}
}
