package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	
)

//go:embed templates
var templateFS embed.FS

const BASE_LAYOUT = "base"

type View struct {
	Template *template.Template
	Layout   string
	Pages    []string
}



var funcs = template.FuncMap{
	"GetYear": func() int {
		return time.Now().Year()
	},
}

func getStaticFiles() []string {
	files, err := filepath.Glob("static/*.*")
	if err != nil {
		panic(err)
	}
	return files
}

func getLayoutFiles() []string {
	files, err := filepath.Glob("templates/*.layout.tmpl")
	if err != nil {
		panic(err)
	}
	return files
}

func addDefaultTemplateData(r *http.Request, data *TemplateData) *TemplateData {
	if data == nil {
		data = &TemplateData{}
	}
	data.User = getUserByCookie(r)
	data.Route = strings.ReplaceAll(r.URL.Path, "/", "")
	if data.Route == ""{
		data.Route = "index"
	}
	return data
}

func NewView(pages ...string) (*View, error) {
	t, err := parseTemplates(pages...)
	if err != nil {
		return nil, err
	}
	return &View{
		Template: t,
		Layout:   BASE_LAYOUT,
		Pages:    pages,
	}, nil
}

func (v *View) Render(w http.ResponseWriter, r *http.Request, data *TemplateData) error {
	if data == nil{
		data = &TemplateData{}
	}

	data = addDefaultTemplateData(r, data)
	if env == "dev" {
		t, err := parseTemplates(v.Pages...)
		if err != nil {
			return err
		}
		v.Template = t
	}

	return v.Template.ExecuteTemplate(w, v.Layout, data)
}

func parseTemplates(pages ...string) (*template.Template, error) {
	files := getLayoutFiles()
	files = append(files, getStaticFiles()...)
	for _, f := range pages {
		files = append(files, fmt.Sprintf("templates/%s.page.tmpl", f))
	}

	var t *template.Template
	var err error

	_, exists := cache[pages[0]]

	if !exists {
		if env == "dev" {
			t, err = template.New("").Funcs(funcs).ParseFiles(files...)
			
		} else {
			t, err = template.New("").Funcs(funcs).ParseFS(templateFS, files...)
			
			cache[pages[0]] = t
		}
	} else {
		t = cache[pages[0]]
	}

	if err != nil {
		return nil, err
	}

	return t, nil
}