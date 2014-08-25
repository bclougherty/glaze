package controllers

import (
	"net/http"
	"github.com/octoberxp/glaze/glaze"
)

type Public struct {
	*core.Controller
}

func NewPublicController() *Public {
	controller, err := core.NewController()
	if err != nil {
		panic(err)
	}

	controller.LoadTemplates(core.TemplatePath + "public/*.html")

	return &Public{Controller: controller}
}

func (controller *Public) Index(w http.ResponseWriter, r *http.Request) {
	controller.RenderTemplate(w, "index", nil)
}
