package model

import "gorm.io/gorm"

//Carousel 轮播图，即轮着播放图片,为了突出某个商品
type Carousel struct {
	gorm.Model
	ImgPath   string
	ProductId uint `gorm:"not null"` //商品
}
