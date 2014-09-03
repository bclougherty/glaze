package glaze

import (
	"fmt"
	"html/template"
	"net/http"
	"path"
	"path/filepath"
)

// Controller structs hold the data necessary to bind view paths. A Controller
// can have any number of http handler methods added to it, and calling
// RenderTemplate from within a controller method will execute the given template.
type Controller struct {
	BaseTemplate *template.Template
	Templates    TemplateMap

	templatePath string
}

// TemplateMap maps string names to parsed html templates
type TemplateMap map[string]*template.Template

// NewController will create a basic controller.
// templatePath is the full path to the template root
// controllerTemplatePath is the relative path from the template root to
//     this controller's template directory
func NewController(templatePath, controllerTemplatePath string) (*Controller, error) {
	tmpl, err := template.ParseFiles(path.Join(templatePath, "layouts/default.html"))
	if err != nil {
		return nil, err
	}

	controller := &Controller{BaseTemplate: tmpl, templatePath: templatePath}

	controller.loadTemplates(controllerTemplatePath)

	return controller, nil
}

// RenderTemplate will execute the named template using the given writer
func (controller *Controller) RenderTemplate(writer http.ResponseWriter, templateName string, data interface{}) {
	err := controller.Templates[templateName+".html"].Execute(writer, data)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (controller *Controller) loadTemplates(templatePath string) {
	if templatePath == "" {
		return
	}

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
