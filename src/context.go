package main

import (
	"fmt"
	"net/http"
)

type Context struct {
	write      http.ResponseWriter
	req        *http.Request
	Path       string
	Method     string
	Params     map[string]string
	StatusCode int
	handlers   []HandlerFunc
	index      int
	engine     *Engine
}

func (c *Context) Status(status int) {
	c.StatusCode = status
}
func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.write.Write(data)
}
func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.write, name, data); err != nil {
		c.Status(500)
	}
}
func (c *Context) SetHeader(key string, value string) {
	c.write.Header().Set(key, value)
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json; charset=utf-8")
	c.Status(code)
}
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain; charset=utf-8")
	c.Status(code)
	c.write.Write([]byte(fmt.Sprintf(format, values...)))
}
func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		write:  w,
		req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
		index:  -1,
	}
}
func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}
