package router

//---------------------------------------------------------------------------------------------------
// execution.go
//---------------------------------------------------------------------------------------------------

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	chi "github.com/go-chi/chi"
)

//SetConfigDefault can be used to set all of the global variable configuration items to their default
// value
func SetConfigDefault() {
	ConfigListenAndServeWait = DefaultListenAndServeWait
	ConfigServerShutdownTimeout = DefaultServerShutdownTimeout
}

//SetConfigFromEnv can be used to read a map of environmental variables, grab the keys and convert
// those string configuration options into values to validate and place into the global variables
func SetConfigFromEnv(envs map[string]string) {
	//get the configuration option for listen and wait serve
	if listenAndWaitServe, ok := envs[EnvNameListeandServeWait]; ok {
		//ensure that its not empty
		if listenAndWaitServe != "" {
			//attempt to convert to an integer
			if listenAndWaitServeInt, err := strconv.Atoi(listenAndWaitServe); err == nil {
				//set integer in milliseconds
				listenAndWaitServeDuration := time.Duration(listenAndWaitServeInt) * time.Millisecond
				//ensure that it's not less or equal to 0
				if listenAndWaitServeDuration <= 0 {
					listenAndWaitServeDuration = DefaultListenAndServeWait
				}
				//set configuration
				ConfigListenAndServeWait = listenAndWaitServeDuration
			}
		}
	}
	//get the configuration option for shutdown timeout
	if serverShutdownTimeout, ok := envs[EnvNameServerShutdownTmeout]; ok {
		//ensure that its not empty
		if serverShutdownTimeout != "" {
			//attempt to convert to an integer
			if serverShutdownTimeoutInt, err := strconv.Atoi(serverShutdownTimeout); err == nil {
				//set integer in milliseconds
				serverShutdownTimeoutDuration := time.Duration(serverShutdownTimeoutInt) * time.Millisecond
				//ensure that it's not less or equal to 0
				if serverShutdownTimeoutDuration <= 0 {
					serverShutdownTimeoutDuration = DefaultServerShutdownTimeout
				}
				//set configuration
				ConfigServerShutdownTimeout = serverShutdownTimeoutDuration
			}
		}
	}

}

//ValidateRoutes can be used to confirm that a slice of provided routes has no common errors
func ValidateRoutes(routes []RouteConfiguration) (err error) {
	var routeMap = make(map[string]struct{})
	var empty struct{}

	//populate map for routes to check for duplicates
	for _, route := range routes {
		//check if handle function is nil
		if route.HandleFx == nil {
			err = errors.New(ErrHandleFxNil)

			return
		}
		//check if method is empty
		if route.Method == "" {
			err = errors.New(ErrMethodEmpty)

			return
		}
		//check if method is valid
		if _, ok := validMethods()[route.Method]; !ok {
			err = fmt.Errorf(ErrInvalidMethod, route.Method)

			return
		}
		//populate route map, be sure to include the method since you can have
		// multiple methods for the same route
		routeMap[route.Route+route.Method] = empty
	}
	//validate routes (check to see if there are any duplicate routes)
	//if len(routeMap) != len(routes) {
	//	err = errors.New(ErrDuplicateRoute)

	//	return
	//}

	return
}

//ValidateHandles can be used to confirm that a slice of provided handles has no common errors
func ValidateHandles(handles []HandleConfiguration) (err error) {
	var handleMap = make(map[string]struct{})
	var empty struct{}

	//populate map for routes to check for duplicates
	for _, handle := range handles {
		//check if handle function is nil
		if handle.HandleFx == nil {
			err = errors.New(ErrHandleFxNil)

			return
		}
		//populate handle map using the route and whether or not there is a pattern
		handleMap[fmt.Sprintf("%s%t", handle.Route, handle.Pattern)] = empty
	}
	//validate routes (check to see if there are any duplicate routes)
	//if len(handleMap) != len(handles) {
	//	err = errors.New(ErrDuplicateHandle)

	//	return
	//}

	return
}

//GetRouteVariable can be used to pull variables out of the given request's route
func GetRouteVariable(request *http.Request, key string) (value string, err error) {
	//get the route context
	ctx := chi.RouteContext(request.Context())
	//get the parameters from the URL
	params := ctx.URLParams
	//range through the keys and try to find the key
	for i, k := range params.Keys {
		if key == k {
			value = params.Values[i]

			return
		}
	}
	//output error since key not found
	err = fmt.Errorf(ErrRouteKeyNotFoundf, key)

	return
}

// SortRoutes sorts routes
func SortRoutes(routes []RouteConfiguration) (sortedRoutes []RouteConfiguration) {
	var routesWithVariables []RouteConfiguration
	var routesWithNoVariables []RouteConfiguration

	//range over the routes
	for _, route := range routes {
		//switch on conditions to determine how to sort
		switch {
		case strings.Contains(route.Route, "{"), strings.Contains(route.Route, "}"):
			//if the route contains route variables, noted by the brackets, sort for bracket
			routesWithVariables = append(routesWithVariables, route)
		default:
			//these are default cases where no curly braces are present
			routesWithNoVariables = append(routesWithNoVariables, route)
		}
	}
	//sort the routes, putting the routes with no variables in front
	sortedRoutes = append(sortedRoutes, routesWithNoVariables...)
	sortedRoutes = append(sortedRoutes, routesWithVariables...)

	return
}

// SortHandles sorts handles
func SortHandles(handles []HandleConfiguration) (sortedHandles []HandleConfiguration) {
	var handlesWithVariables []HandleConfiguration
	var handlesWithVariablesAndPatterns []HandleConfiguration
	var handlesWithPattern []HandleConfiguration
	var handlesWithout []HandleConfiguration

	//range over the routes
	for _, handle := range handles {
		//switch on conditions to determine how to sort
		switch {
		case (strings.Contains(handle.Route, "{") || strings.Contains(handle.Route, "}")) && !handle.Pattern:
			handlesWithVariables = append(handlesWithVariables, handle)
		case (strings.Contains(handle.Route, "{") || strings.Contains(handle.Route, "}")) && handle.Pattern:
			handlesWithVariablesAndPatterns = append(handlesWithVariablesAndPatterns, handle)
		case !(strings.Contains(handle.Route, "{") || strings.Contains(handle.Route, "}")) && handle.Pattern:
			handlesWithPattern = append(handlesWithPattern, handle)
		default:
			handlesWithout = append(handlesWithout, handle)
		}
	}
	//sort the handles
	sortedHandles = append(sortedHandles, handlesWithout...)
	sortedHandles = append(sortedHandles, handlesWithVariables...)
	sortedHandles = append(sortedHandles, handlesWithPattern...)
	sortedHandles = append(sortedHandles, handlesWithVariablesAndPatterns...)

	return
}
