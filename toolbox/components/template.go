package template

import (
	htmpl "html/template"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cosiner/zerver"
)

type (
	TemplateOption struct {
		DelimLeft  string
		DelimRight string
		Suffixes   []string
		Path       []string
		htmpl.FuncMap
	}
	// Template is a template engine which support set delimiters, add file suffix name,
	// add template file and dirs, add template functions, compile template, and
	// render template
	Template interface {
		zerver.Component
		Lookup(name string) Template
		Render(w io.Writer, data interface{}) error
		RenderTmpl(w io.Writer, name string, data interface{}) error
	}

	// template implements Template interface use standard html/template package
	template htmpl.Template
)

func (t *TemplateOption) init() {
	if t.DelimLeft == "" || t.DelimRight == "" {
		t.DelimLeft, t.DelimRight = "{{", "}}"
	}
	if t.Suffixes == nil {
		t.Suffixes = []string{"html", "tmpl"}
	}
	if t.Path == nil {
		t.Path = []string{"views"}
	}
}

func NewTemplate(o *TemplateOption) (Template, error) {
	if o == nil {
		o = &TemplateOption{}
	}
	o.init()
	files, err := filenames(o.Path, o.Suffixes)
	if err == nil {
		tmpl, e := htmpl.New("tmpl").Delims(o.DelimLeft, o.DelimRight).Funcs(o.FuncMap).ParseFiles(files...)
		if e == nil {
			return (*template)(tmpl), nil
		}
		err = e
	}
	return nil, err
}

func (t *template) Init(zerver.Enviroment) error {
	return nil
}

func (t *template) Destroy() {}

func (t *template) Render(w io.Writer, data interface{}) error {
	return (*htmpl.Template)(t).Execute(w, data)
}

func (t *template) RenderTmpl(w io.Writer, name string, values interface{}) error {
	return (*htmpl.Template)(t).ExecuteTemplate(w, name, values)
}

func (t *template) Lookup(name string) Template {
	return (*template)((*htmpl.Template)(t).Lookup(name))
}

func filenames(path []string, suffixes []string) (files []string, err error) {
	suffMap := make(map[string]bool)
	for _, s := range suffixes {
		suffMap[s] = true
	}
	addTmpl := func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && isTemplate(path, suffMap) {
			files = append(files, path)
		}
		return err
	}
	for _, p := range path {
		if err = filepath.Walk(p, addTmpl); err != nil {
			break
		}
	}
	return files, err
}

// isTemplate check whether a file name is recognized template file
func isTemplate(name string, suffixes map[string]bool) (is bool) {
	index := strings.LastIndex(name, ".")
	if is = (index >= 0); is {
		is = suffixes[name[index+1:]]
	}
	return
}
