package gee

import (
	"net/http"
)

// HandlerFunc 定义gee使用的请求处理器
type HandlerFunc func(c *Context)

// Engine 用于实现接口ServeHTTP
// router 是路由映射表
// key 由请求方法和静态路由地址构成, value 是用户映射的处理方法
type Engine struct {
	router *router
}

// New 是gee.Engine的构造器
func New() *Engine {
	return &Engine{router: newRouter()}
}

func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	engine.router.addRoute(method, pattern, handler)
}

// Get 定义添加GET请求的方法
func (engine *Engine) Get(pattern string, handler HandlerFunc) {
	engine.addRoute("GET", pattern, handler)
}

// POST 定义添加POST请求的方法
func (engine *Engine) POST(pattern string, handler HandlerFunc) {
	engine.addRoute("POST", pattern, handler)
}

// Run 定义启动http服务的方法
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := NewContext(w, r)
	engine.router.handle(c)
}
