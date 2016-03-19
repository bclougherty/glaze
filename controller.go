package glaze

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"path/filepath"

	"github.com/oxtoacart/bpool"
)

var bufpool *bpool.BufferPool

// Controller is a convenience interface that allows classes with an embedded BaseController
// to be passed to various glaze methods.
type Controller interface {
	RenderHTML(writer http.ResponseWriter, templateName string, data interface{}) error
	RenderJSON(writer http.ResponseWriter, data interface{}) error
}

// BaseController structs hold the data necessary to bind view paths. A Controller
// can have any number of http handler methods added to it, and calling
// RenderTemplate from within a controller method will execute the given template.
type BaseController struct {
	BaseTemplate *template.Template
	Templates    TemplateMap

	templatePath string
}

// TemplateMap maps string names to parsed html templates
type TemplateMap map[string]*template.Template

func init() {
	bufpool = bpool.NewBufferPool(64)
}

// NewController will create a basic controller.
// templatePath is the full path to the template root
// controllerTemplatePath is the relative path from the template root to
//     this controller's template directory
func NewController(templatePath, controllerTemplatePath string, funcMap template.FuncMap) (*BaseController, error) {
	controller := &BaseController{templatePath: templatePath}
	controller.loadTemplates(controllerTemplatePath, funcMap)

	return controller, nil
}

// RenderHTML will execute the named template using the given writer
func (controller *BaseController) RenderHTML(writer http.ResponseWriter, templateName string, data interface{}) error {
	// Ensure the template exists in the map.
	_, ok := controller.Templates[templateName+".html"]
	if !ok {
		return fmt.Errorf("The template %s does not exist.", templateName)
	}

	// Create a buffer to temporarily write to and check if any errors were encounted.
	buf := bufpool.Get()
	defer bufpool.Put(buf)

	err := controller.Templates[templateName+".html"].ExecuteTemplate(buf, "layout", data)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return err
	}

	_, err = buf.WriteTo(writer)
	if err != nil {
		return err
	}

	return nil
}

// RenderJSON will output the contents of data as JSON, setting an appropriate Content-Type header
func (controller *BaseController) RenderJSON(writer http.ResponseWriter, data interface{}) error {
	// Create a buffer to temporarily write to and check if any errors were encounted.
	buf := bufpool.Get()
	defer bufpool.Put(buf)

	enc := json.NewEncoder(buf)

	if err := enc.Encode(&data); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return err
	}

	writer.Header().Set("Content-Type", "application/json")

	_, err := buf.WriteTo(writer)
	if err != nil {
		return err
	}

	return nil
}

func (controller *BaseController) loadTemplates(templatePath string, funcMap template.FuncMap) error {
	if templatePath == "" {
		return fmt.Errorf("Attempted to load templates from an empty path")
	}

	fullpath := path.Join(controller.templatePath, templatePath, "*.html")

	// fmt.Printf("Looking for layouts in %s...", path.Join(controller.templatePath, "layouts", "*.html"))

	layouts, err := filepath.Glob(path.Join(controller.templatePath, "layouts", "*.html"))
	if err != nil {
		return err
	}

	// fmt.Printf(" found %d\n", len(layouts))

	// fmt.Printf("Looking for templates in %s...", fullpath)

	includes, err := filepath.Glob(fullpath)
	if err != nil {
		return err
	}

	// fmt.Printf(" found %d\n", len(includes))

	templates := make(map[string]*template.Template, len(includes))

	for _, layout := range layouts {
		// fmt.Println(layout)
		for _, file := range includes {
			name := filepath.Base(file)
			// fmt.Printf("\t%s\n", name)
			templates[name] = template.Must(template.New(name).Funcs(funcMap).ParseFiles(file, layout))
		}
	}

	controller.Templates = templates

	return nil
}
