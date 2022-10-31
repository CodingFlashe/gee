package main

import (
	"fmt"
	"gee"
	"html/template"
	"net/http"
	"time"
)

/*

func onlyForV2() gee.HandlerFunc {
	return func(c *gee.Context) {
		t := time.Now()
		c.Fail(500, "Internal Server Error")
		log.Printf("[%d] %s in %v for group v2", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}
*/

type student struct {
	Name string
	Age  int8
}

func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

func main() {
	r := gee.New()
	r.Use(gee.Logger())
	r.SetFuncMap(template.FuncMap{
		"FormatAsDate": FormatAsDate,
	})
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./static")
	stu1 := &student{Name: "Geektutu", Age: 20}
	stu2 := &student{Name: "Jack", Age: 22}
	r.GET("/", func(c *gee.Context) {
		c.HTML(http.StatusOK, "css.tmpl", nil)
	})
	r.GET("/students", func(c *gee.Context) {
		c.HTML(http.StatusOK, "arr.tmpl", gee.H{
			"title":  "gee",
			"stuArr": [2]*student{stu1, stu2},
		})
	})
	r.GET("/date", func(c *gee.Context) {
		c.HTML(http.StatusOK, "custom_func.tmpl", gee.H{
			"title": "gee",
			"now":   time.Date(2019, 8, 17, 0, 0, 0, 0, time.UTC),
		})
	})
	/*
		r.GET("/", func(c *gee.Context) {
			c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
		})
		r.GET("/index", func(c *gee.Context) {
			c.HTML(http.StatusOK, "<h1>Index Page</h1>")
		})
		v1 := r.Group("/v1")
		{
			v1.GET("/", func(c *gee.Context) {
				c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
			})
			v1.GET("/hello", func(c *gee.Context) {
				c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
			})
		}

		v2 := r.Group("/v2")
		v2.Use(onlyForV2())
		{
			v2.GET("/hello/:name", func(c *gee.Context) {
				c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
			})
				v2.POST("/login", func(c *gee.Context) {
					c.JSON(http.StatusOK, gee.H{
						"username": c.PostForm("username"),
						"password": c.PostForm("password"),
					})
				})
		}

		r.GET("/hello/:name/doc", func(c *gee.Context) {
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
		r.GET("/assets/*filepath/oop", func(c *gee.Context) {
			c.JSON(http.StatusOK, gee.H{
				"filepath": c.Param("filepath"),
			})
		})
	*/
	r1 := gee.Default()
	r1.GET("/", func(c *gee.Context) {
		c.String(http.StatusOK, "Hello Geektutu\n")
	})
	r1.GET("/panic", func(c *gee.Context) {
		names := []string{"geektutu"}
		c.String(http.StatusOK, names[100])
	})
	r1.Run(":9999")
}
