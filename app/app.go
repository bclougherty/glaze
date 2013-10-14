package main

import (
	"github.com/octoberxp/glaze/controllers"
	"net/http"
)

func main() {
	// Instantiate controllers
	public := controllers.NewPublicController()

	// Configure routes
	// TODO: Move this somewhere a little more sensible
	http.HandleFunc("/", public.Index)

	// And start the server
	http.ListenAndServe(":9000", nil)
}
