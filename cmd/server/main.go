package main

import (
	"html/template"

	"github.com/labstack/echo/v4"
	"github.com/palo-verde-digital/codex/internal/handler"
	"github.com/palo-verde-digital/codex/internal/render"
)

func main() {
	r := &render.Renderer{
		Templates: template.Must(template.ParseGlob("dist/views/*.html")),
	}

	e := echo.New()
	e.Renderer = r
	e.Static("/dist", "dist")

	e.GET("/", handler.Index)

	e.Start(":8080")
}
