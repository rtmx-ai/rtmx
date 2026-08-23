// Package dashboard provides embedded web assets for the RTMX dashboard SPA.
// All static files (HTML, CSS, JS) are compiled into the Go binary via embed.FS.
// No Node.js or JavaScript bundler is required.
package dashboard

import (
	"embed"
	"html/template"
	"io"
	"io/fs"
	"net/http"
)

//go:embed static templates
var assets embed.FS

// Assets returns the embedded filesystem containing all dashboard assets.
func Assets() fs.FS {
	return assets
}

// StaticFS returns an http.FileSystem for serving static assets.
func StaticFS() http.FileSystem {
	sub, _ := fs.Sub(assets, "static")
	return http.FS(sub)
}

// TemplateFS returns the embedded templates sub-filesystem.
func TemplateFS() fs.FS {
	sub, _ := fs.Sub(assets, "templates")
	return sub
}

// HTMLContent is an alias for template.HTML to allow safe HTML injection.
type HTMLContent = template.HTML

// LayoutData holds the data passed to the layout template.
type LayoutData struct {
	Title      string
	ActivePage string
	Content    HTMLContent
}

// RenderLayout renders the base layout template with the given data.
func RenderLayout(w io.Writer, data LayoutData) error {
	tmpl, err := template.ParseFS(assets, "templates/layout.html")
	if err != nil {
		return err
	}
	return tmpl.Execute(w, data)
}

// RenderPartial renders a named partial template.
func RenderPartial(w io.Writer, name string, data interface{}) error {
	path := "templates/partials/" + name + ".html"
	tmpl, err := template.ParseFS(assets, path)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, data)
}
