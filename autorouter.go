package glaze

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/octoberxp/go-utils/stringutils"
)

// GenerateRoutes calls GenerateRoutesWithPrefix, using the spinal-cased controller class name as the prefix.
func GenerateRoutes(controller interface{}) map[string]func(w http.ResponseWriter, r *http.Request) {
	structType := reflect.TypeOf(controller)
	controllerName := structType.Name()

	return GenerateRoutesWithPrefix(controller, stringutils.CamelToSpinal(controllerName))
}

// GenerateRoutesWithPrefix generates a map of urls to controller methods, suitable for passing off to http.Handle.
// It will create one route per public method of controller that is not inherited from GlazeController.
// Each route will be in the form "/[prefix]/[method-name]", where prefix is preserved exactly, and method names are
// converted to spinal-case.
func GenerateRoutesWithPrefix(controller interface{}, prefix string) map[string]func(w http.ResponseWriter, r *http.Request) {
	// create a Glaze controller and get a list of its methods
	// so that we can exclude them from the list of handle-able methods
	glazeController := &Controller{}

	structType := reflect.TypeOf(glazeController)
	numberOfMethods := structType.NumMethod()

	var glazeControllerMethods []string

	for i := 0; i < numberOfMethods; i++ {
		glazeControllerMethods = append(glazeControllerMethods, structType.Method(i).Name)
	}

	routes := make(map[string]func(w http.ResponseWriter, r *http.Request))

	structType = reflect.TypeOf(controller)
	numberOfMethods = structType.NumMethod()

	requestType := reflect.TypeOf(&http.Request{})

	for i := 0; i < numberOfMethods; i++ {
		method := structType.Method(i)

		// if this is a glaze controller method, then we're not interested
		if contains(glazeControllerMethods, method.Name) {
			continue
		}

		// make sure this is a http handler func - the first argument will be the controller pointer
		if method.Type.NumIn() != 3 || method.Type.In(2) != requestType {
			continue
		}

		routePath := fmt.Sprintf("/%s/%s", prefix, stringutils.CamelToSpinal(method.Name))

		routes[routePath] = func(w http.ResponseWriter, r *http.Request) {
			reflect.ValueOf(controller).MethodByName(method.Name).Call([]reflect.Value{
				reflect.ValueOf((w)),
				reflect.ValueOf((r)),
			})
		}
	}

	return routes
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
