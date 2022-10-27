package gee

import (
	"log"
	"net/http"
)

// HandlerFunc 定义gee使用的请求处理器
type HandlerFunc func(c *Context)

// Engine 用于实现接口ServeHTTP
// router 是路由映射表
// key 由请求方法和静态路由地址构成, value 是用户映射的处理方法
type (
	RouterGroup struct {
		prefix      string
		middlewares []HandlerFunc //支持中间件
		parent      *RouterGroup  //支持嵌套
		engine      *Engine       //所有分组共享同一个engine实例
	}
	Engine struct {
		*RouterGroup
		router *router
		groups []*RouterGroup
	}
)

// New 是gee.Engine的构造器
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

// Group 被定义用来产生一个新的分组
// 记住所有的组都共享相同的engine实例
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

// GET 定义添加GET请求的方法
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

// POST 定义添加POST请求的方法
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

// Run 定义启动http服务的方法
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := newContext(w, r)
	engine.router.handle(c)
}
