package model

// BasePage 分页操作
type BasePage struct {
	PageNum  int `form:"pageNum"`
	PageSize int `form:"pageSize"`
}
