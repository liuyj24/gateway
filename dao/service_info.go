package dao

import (
	"time"

	"github.com/liuyj/gateway/public"

	"github.com/liuyj/gateway/dto"

	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
)

type ServiceInfo struct {
	ID          int64     `json:"id" gorm:"primary_key"`
	LoadType    int       `json:"load_type" gorm:"column:load_type" description:"负载类型 0=http 1=tcp 2=grpc"`
	ServiceName string    `json:"service_name" gorm:"column:service_name" description:"服务名称"`
	ServiceDesc string    `json:"service_desc" gorm:"column:service_desc" description:"服务描述"`
	UpdatedAt   time.Time `json:"create_at" gorm:"column:create_at" description:"更新时间"`
	CreatedAt   time.Time `json:"update_at" gorm:"column:update_at" description:"添加时间"`
	IsDelete    int8      `json:"is_delete" gorm:"column:is_delete" description:"是否已删除；0：否；1：是"`
}

func (si *ServiceInfo) TableName() string {
	return "gateway_service_info"
}

func (si *ServiceInfo) ServiceDetail(c *gin.Context, db *gorm.DB, serviceInfo *ServiceInfo) (*ServiceDetail, error) {
	if serviceInfo.ServiceName == "" {
		info, err := si.Find(c, db, serviceInfo)
		if err != nil {
			return nil, err
		}
		serviceInfo = info
	}
	httpRule := &HttpRule{ServiceID: serviceInfo.ID}
	httpRule, err := httpRule.Find(c, db, httpRule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	tcpRule := &TcpRule{ServiceID: serviceInfo.ID}
	tcpRule, err = tcpRule.Find(c, db, tcpRule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	grpcRule := &GrpcRule{ServiceID: serviceInfo.ID}
	grpcRule, err = grpcRule.Find(c, db, grpcRule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	accessControl := &AccessControl{ServiceID: serviceInfo.ID}
	accessControl, err = accessControl.Find(c, db, accessControl)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	loadBalance := &LoadBalance{ServiceID: serviceInfo.ID}
	loadBalance, err = loadBalance.Find(c, db, loadBalance)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	detail := &ServiceDetail{
		Info:          serviceInfo,
		HTTPRule:      httpRule,
		TCPRule:       tcpRule,
		GRPCRule:      grpcRule,
		LoadBalance:   loadBalance,
		AccessControl: accessControl,
	}
	return detail, nil
}

func (si *ServiceInfo) PageList(c *gin.Context, db *gorm.DB, param *dto.ServiceListInput) ([]ServiceInfo, int64, error) {
	var total int64
	var list []ServiceInfo
	offset := (param.PageNo - 1) * param.PageSize

	query := db.SetCtx(public.GetGinTraceContext(c))
	query = query.Table(si.TableName()).Where("is_delete=0")

	//add search param
	if param.Info != "" {
		query.Where("(service_name like ? or service_desc like ?)", "%"+param.Info+"%", "%"+param.Info+"%")
	}
	if err := query.Limit(param.PageSize).Offset(offset).Order("id desc").Find(&list).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, 0, err
	}
	query.Limit(param.PageSize).Offset(offset).Count(&total)
	return list, total, nil
}

func (*ServiceInfo) Find(c *gin.Context, db *gorm.DB, serviceInfo *ServiceInfo) (*ServiceInfo, error) {
	out := &ServiceInfo{}
	err := db.SetCtx(public.GetGinTraceContext(c)).Where(serviceInfo).Find(out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (si *ServiceInfo) Save(c *gin.Context, db *gorm.DB) error {
	return db.SetCtx(public.GetGinTraceContext(c)).Save(si).Error
}

func (t *ServiceInfo) GroupByLoadType(c *gin.Context, tx *gorm.DB) ([]dto.DashServiceStatItemOutput, error) {
	list := []dto.DashServiceStatItemOutput{}
	query := tx.SetCtx(public.GetGinTraceContext(c))
	if err := query.Table(t.TableName()).Where("is_delete=0").Select("load_type, count(*) as value").Group("load_type").Scan(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
