package component

import (
	tmpl "html/template"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cosiner/zerver"
)

const (
	TEMPLATE = "Template"
)

type (
	TemplateOption struct {
		DelimLeft  string
		DelimRight string
		Suffixes   []string
		Path       []string
		tmpl.FuncMap
	}

	// template implements Template interface use standard html/template package
	Template tmpl.Template
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

func NewTemplate() *Template {
	return (*Template)(tmpl.New("_tmpl_"))
}

func (t *Template) Init(env zerver.Environment) error {
	compEnv := env.(zerver.ComponentEnvironment)
	var o *TemplateOption
	if op := compEnv.GetSetAttr(TEMPLATE, nil); op == nil {
		o = &TemplateOption{}
	} else {
		o = op.(*TemplateOption)
	}
	o.init()

	files, err := filenames(o.Path, o.Suffixes)
	if err == nil {
		_, err = (*tmpl.Template)(t).
			Delims(o.DelimLeft, o.DelimRight).
			Funcs(o.FuncMap).
			ParseFiles(files...)
	}

	return err
}

func (t *Template) Destroy() {}

func (t *Template) Render(w io.Writer, data interface{}) error {
	return (*tmpl.Template)(t).Execute(w, data)
}

func (t *Template) RenderTmpl(w io.Writer, name string, values interface{}) error {
	return (*tmpl.Template)(t).ExecuteTemplate(w, name, values)
}

func (t *Template) Lookup(name string) *Template {
	return (*Template)((*tmpl.Template)(t).Lookup(name))
}

func (t *Template) Template() *tmpl.Template {
	return (*tmpl.Template)(t)
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
