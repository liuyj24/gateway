package dao

import (
	"time"

	"github.com/pkg/errors"

	"github.com/liuyj/gateway/public"

	"github.com/liuyj/gateway/dto"

	"github.com/e421083458/gorm"

	"github.com/gin-gonic/gin"
)

type Admin struct {
	Id        int       `json:"id" gorm:"primary_key" description:"自增主键"`
	UserName  string    `json:"user_name" gorm:"column:user_name" description:"管理员用户名"`
	Salt      string    `json:"salt" gorm:"column:salt" description:"盐"`
	Password  string    `json:"password" gorm:"column:password" description:"密码"`
	UpdatedAt time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	CreatedAt time.Time `json:"create_at" gorm:"column:create_at" description:"创建时间"`
	IsDelete  int       `json:"is_delete" gorm:"column:is_delete" description:"是否删除"`
}

func (t *Admin) TableName() string {
	return "gateway_admin"
}

func (admin *Admin) LoginCheck(c *gin.Context, db *gorm.DB, param *dto.AdminLoginInput) (*Admin, error) {
	result, err := admin.Find(c, db, &Admin{UserName: param.UserName, IsDelete: 0})
	if err != nil {
		//return nil, errors.New("用户信息不存在")
		return nil, err
	}
	saltPassword := public.GenerateSaltPassword(result.Salt, param.Password)
	if saltPassword != result.Password {
		return nil, errors.New("密码错误，请重新输入")
	}
	return result, nil
}

//get admin info from db
func (admin *Admin) Find(c *gin.Context, db *gorm.DB, adminParam *Admin) (*Admin, error) {
	out := &Admin{}
	err := db.SetCtx(public.GetGinTraceContext(c)).Where(adminParam).Find(out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

//save admin info into db
func (admin *Admin) Save(c *gin.Context, db *gorm.DB) error {
	return db.SetCtx(public.GetGinTraceContext(c)).Save(admin).Error
}
