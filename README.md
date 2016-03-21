glaze
=====

A mini web framework in Go. Glaze currently only provides a simple
controller/view abstraction to make it easier to get a simple web
app up and running.

This example assumes that you have a view directory laid out like so:

```
/path/to/views/
	layouts/
		default.html
	public/
		index.html
```

Glaze will always try to load any layout files from "layouts/*.html"
under the given view root. The layout should be inside a {{define "layout"}}...{{end}} block.


controllers/public.go:

```go
package controllers

import (
	"net/http"

	"github.com/octoberxp/glaze"
)

type Public struct {
	*glaze.Controller
}

func NewPublicController() *Public {
	controller, err := glaze.NewController("/path/to/views/", "public")
	if err != nil {
		panic(err)
	}

	return &Public{Controller: controller}
}

func (controller *Public) Index(w http.ResponseWriter, r *http.Request) {
	controller.RenderTemplate(w, "index", nil)
}
```

main.go:

```go
package main

import (
	"net/http"
	
	"exampleapp/controllers"
)

func main() {
	// Instantiate controllers
	public := controllers.NewPublicController()

	// Configure routes
	http.HandleFunc("/", public.Index)

	// And start the server
	http.ListenAndServe(":9000", nil)
}
```