package router

import (
	"net/http"
	"time"
)

//---------------------------------------------------------------------------------------------------
// types.go
//---------------------------------------------------------------------------------------------------

//common constants used throughout the package
const (
	LogPrefix string = "Router: "
)

//error constants
const (
	ErrDuplicateRoute   string = "Duplicate routes found"
	ErrDuplicateHandle  string = "Duplicate handles found"
	ErrRouterStarted    string = "Router already started"
	ErrRouterNotStarted string = "Router not started"
)

//configuration constants
var (
	ConfigListenAndServeWait    = DefaultListenAndServeWait    //how long to block after listen and serve
	ConfigServerShutdownTimeout = DefaultServerShutdownTimeout //how long to wait for a shutdown to occur
)

//default configuration constants
const (
	DefaultListenAndServeWait    time.Duration = 1 * time.Second  //default for ConfigListenAndServeWait
	DefaultServerShutdownTimeout time.Duration = 10 * time.Second //default for ConfigServerShutdownTimeout
)

//RouteConfiguration provides a struct that can be used to configure a route
type RouteConfiguration struct {
	Route    string                                   //route to listen to
	Method   string                                   //method for the provided route
	HandleFx func(http.ResponseWriter, *http.Request) //handle function to execute on receiving request
}

//HandleConfiguration provides a struct that can be used to configure a handle
type HandleConfiguration struct {
	Route    string       //route to listen to
	Pattern  bool         //whether or not the configuration is a pattern
	HandleFx http.Handler //handle function
}

//validMethods returns a map of valid methods for http endpoints
func validMethods() map[string]string {
	return map[string]string{
		"GET":     "POST",
		"PUT":     "PUT",
		"POST":    "POST",
		"PATCH":   "PATCH",
		"DELETE":  "DELETE",
		"OPTIONS": "OPTIONS",
		"HEAD":    "HEAD",
		"TRACE":   "TRACE",
		"CONNECT": "CONNECT",
	}
}
