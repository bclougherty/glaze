package glaze

import (
	"fmt"
	"html/template"
	"net/http"
	"path"
	"path/filepath"
)


type Controller struct {
	BaseTemplate *template.Template
	Templates    TemplateMap

	templatePath string
}

type TemplateMap map[string]*template.Template

func NewController(templatePath string) (*Controller, error) {
	tmpl, err := template.ParseFiles(path.Join(templatePath, "layouts/default.html"))
	if err != nil {
		return nil, err
	}

	return &Controller{BaseTemplate: tmpl, templatePath: templatePath}, nil
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
	fullpath := path.Join(controller.templatePath, templatePath, "*.html")

	fmt.Printf("Loading templates from \"%s\"... ", fullpath)

	files, err := filepath.Glob(fullpath)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d\n", len(files))

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
