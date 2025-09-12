package render

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

type Renderer struct {
	Templates *template.Template
}

func (r *Renderer) Render(w io.Writer, name string, data any, c echo.Context) error {
	return r.Templates.ExecuteTemplate(w, name, data)
}
