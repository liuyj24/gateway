package controller

import (
	"fmt"
	"net/http"

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
	group.GET("/service_list", service.serviceList)
	group.POST("/service_delete", service.ServiceDelete)
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
func (service *ServiceController) serviceList(c *gin.Context) {
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
