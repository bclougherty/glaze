package core

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

type TemplateMap map[string]*template.Template

type Controller struct {
	BaseTemplate *template.Template
	Templates    TemplateMap
}

var TemplatePath = os.ExpandEnv("$GOPATH/src/github.com/octoberxp/glaze/views/")

func NewController() (*Controller, error) {
	tmpl, err := template.ParseFiles(TemplatePath + "main/base.html")
	if err != nil {
		return nil, err
	}

	return &Controller{BaseTemplate: tmpl}, nil
}

func (controller *Controller) RenderPage(writer http.ResponseWriter, templateName string, data interface{}) {
	err := controller.Templates[templateName+".html"].Execute(writer, data)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (controller *Controller) RenderTemplate(writer http.ResponseWriter, templateName string, data interface{}) {
	err := controller.Templates[templateName+".html"].Execute(writer, data)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (controller *Controller) LoadTemplates(templatePath string) {
	files, err := filepath.Glob(templatePath)
	if err != nil {
		panic(err)
	}

	templates := make(map[string]*template.Template, len(files))
	for _, file := range files {
		thisTemplate, err := controller.BaseTemplate.Clone()
		thisTemplate, err = thisTemplate.ParseFiles(file)
		if err != nil {
			panic(err)
		}

		templates[filepath.Base(file)] = thisTemplate
	}

	controller.Templates = templates
}
