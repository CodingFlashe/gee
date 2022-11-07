package model

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	Name          string
	Category      uint
	Title         string //商品标题
	Info          string //商品信息
	ImgPath       string //商品展示图
	Price         string
	DiscountPrice string //折后价
	OnSale        bool   //商品是否在售
}
