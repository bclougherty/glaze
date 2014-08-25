package controllers

import (
	"net/http"
	"os"
	"github.com/octoberxp/glaze/glaze"
)

type Public struct {
	*glaze.Controller
}

func NewPublicController() *Public {
	controller, err := glaze.NewController(os.ExpandEnv("./views/"))
	if err != nil {
		panic(err)
	}

	controller.LoadTemplates("public/*.html")

	return &Public{Controller: controller}
}

func (controller *Public) Index(w http.ResponseWriter, r *http.Request) {
	controller.RenderTemplate(w, "index", nil)
}
