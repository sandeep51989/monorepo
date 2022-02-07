package router

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
