package model

import "gorm.io/gorm"

// Cart 购物车模型
type Cart struct {
	gorm.Model
	UserId    uint `gorm:"not null"` //用户
	ProductId uint `gorm:"not null"` //产品
	BossId    uint `gorm:"not null"` //商家
	Num       uint `gorm:"not null"` //数量
	MaxNum    uint `gorm:"not null"` //商品数量限额
	check     bool //检查是否支付
}
