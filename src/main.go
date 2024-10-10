package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

type Engine struct {
	router *router
	*RouterGroup
	groups        []*RouterGroup
	htmlTemplates *template.Template
	funcMap       template.FuncMap
}
type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc
	parent      *RouterGroup
	engine      *Engine
}

func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		engine: engine,
		parent: group,
		prefix: group.prefix + prefix,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

type HandlerFunc func(c *Context)

func (group *RouterGroup) addroute(method string, comp string, handler HandlerFunc) {
	pattern := comp + group.prefix
	log.Printf("addrouter: comp=%s, method=%s, comp=%s", comp, method, pattern)
	group.engine.router.addRoute(method, comp, handler)
}

func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		// Check if file exists and/or if we have permission to access it
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		fileServer.ServeHTTP(c.write, c.req)
	}
}
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlpattern := path.Join(relativePath, "/*filepath")
	group.addroute("GET", urlpattern, handler)
}
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	c.handlers = middlewares
	c.engine = engine
	engine.router.handle(c)
}
func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	method := req.Method
	path := req.URL.Path
	n, params := r.getRoute(method, path)
	if n != nil {
		key := method + "-" + n.pattern
		if handler, ok := r.handlers[key]; ok {
			ctx := &Context{
				write:      w,
				req:        req,
				Path:       path,
				Method:     method,
				Params:     params,
				StatusCode: http.StatusOK,
			}
			handler(ctx)
		}
	} else {
		http.NotFound(w, req)
	}
}
func (group *RouterGroup) Use(middleware ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middleware...)
}
func (group *RouterGroup) GET(path string, handler HandlerFunc) {
	group.addroute("GET", path, handler)
}
func (group *RouterGroup) POST(path string, handler HandlerFunc) {
	group.addroute("POST", path, handler)
}
func (group *RouterGroup) PUT(path string, handler HandlerFunc) {
	group.addroute("PUT", path, handler)
}
func (group *RouterGroup) DELETE(path string, handler HandlerFunc) {
	group.addroute("DELETE", path, handler)
}
func (group *RouterGroup) PATCH(path string, handler HandlerFunc) {
	group.addroute("PATCH", path, handler)
}
func (group *RouterGroup) HEAD(path string, handler HandlerFunc) {
	group.addroute("HEAD", path, handler)
}
func (group *RouterGroup) OPTIONS(path string, handler HandlerFunc) {
	group.addroute("OPTIONS", path, handler)
}
func main() {
	router := newRouter()
	router.addRoute(http.MethodGet, "/b", func(c *Context) {
		fmt.Fprintf(c.write, "11111!")
	})
	router.addRoute(http.MethodPost, "/a", func(c *Context) {
		fmt.Fprintf(c.write, "222222!")
	})
	router.addRoute(http.MethodGet, "/p/:lang/:type", func(c *Context) {
		fmt.Fprintf(c.write, "Lang: %s, Type: %s", c.Param("lang"), c.Param("type"))
	})
	log.Fatal(http.ListenAndServe(":8080", router))
}
