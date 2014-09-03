package glaze

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/octoberxp/go-utils/stringutils"
)

// GenerateRoutes generates a map of urls to controller methods, suitable for passing off to http.Handle.
// This is still in progress - the idea is to do all the reflection at load time and pass back a simple
// map of route to function. It's not quite there yet, because I haven't figured out the last piece.
func GenerateRoutes(controller interface{}) map[string]func(w http.ResponseWriter, r *http.Request) {
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
	controllerName := structType.Elem().Name()
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

		routePath := fmt.Sprintf("/%s/%s", stringutils.CamelToSpinal(controllerName), stringutils.CamelToSpinal(method.Name))

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
