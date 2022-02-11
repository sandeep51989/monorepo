package router

//---------------------------------------------------------------------------------------------------
// router.go
//---------------------------------------------------------------------------------------------------

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	chi "github.com/go-chi/chi"
)

//ensure that logger implements the Logger interface
var (
	_ Router = &router{}
)

//Logger interfacce
type Router interface {
	Start(addr, port string, routes []RouteConfiguration, handles []HandleConfiguration)
	Stop() error
	SetMiddleware(middlewares ...func(http.Handler) http.Handler) error
	SetSecurity(certfile, keyfile string, tlsConfig *tls.Config) error
	Close() error
	LaunchServer()
}

//Router provide a struct that can house the http rest server, the mux router and be used
// for pointer methods for the struct (and as a result have access to the registry)
type router struct {
	sync.RWMutex                                     //read/write mutex
	sync.WaitGroup                                   //waitgroup to track active goRoutines
	started        bool                              //whether or not the router has started
	router         *chi.Mux                          //router for the http.Server
	restServer     *http.Server                      //the rest server
	log            *log.Logger                       //logger to print to
	routes         map[string]RouteConfiguration     //routes that have been added
	handles        map[string]HandleConfiguration    //handles that have been added
	usingSecurity  bool                              //Flag to check for security
	certfile       string                            //Path to cert file
	keyfile        string                            //Path to key file
	tlsConfig      *tls.Config                       //TLS configuration for when using security
	middlewares    []func(http.Handler) http.Handler //middlewares used by the router
}

//NewRouter will create a pointer to an endpoints struct and create all of its internal pointers
func NewRouter(log *log.Logger) *Router {
	routes := make(map[string]RouteConfiguration)
	handles := make(map[string]HandleConfiguration)

	return &Router{
		log:     log,
		routes:  routes,
		handles: handles,
		started: false,
	}
}

//SetSecurity is used to set the security of the server and check the certs used
func (r *Router) SetSecurity(certfile, keyfile string, tlsConfig *tls.Config) (err error) {
	r.Lock()
	defer r.Unlock()
	//we re-create the pointers here, so we can "re-use" the router pointer if necessary
	if r.started {
		err = errors.New(ErrRouterStarted)
		return
	}
	//Check the certs
	_, err = tls.LoadX509KeyPair(certfile, keyfile)
	if err != nil {
		return err
	}
	r.usingSecurity = true
	r.certfile = certfile
	r.keyfile = keyfile
	//TLS Config can be empty
	r.tlsConfig = tlsConfig
	return
}

//SetMiddleware allows the user to create own or pass chi middeware into router
func (r *Router) SetMiddleware(middlewares ...func(http.Handler) http.Handler) (err error) {
	r.middlewares = append(r.middlewares, middlewares...)
	return
}

//Start can be used to start the rest server and provide an error if the service encounters an error
func (r *Router) Start(addr, port string, routes []RouteConfiguration, handles []HandleConfiguration) (err error) {
	r.Lock()
	defer r.Unlock()

	var configListenAndServeWait time.Duration

	//we re-create the pointers here, so we can "re-use" the router pointer if necessary
	if r.started {
		err = errors.New(ErrRouterStarted)
		return
	}
	//validate and set configuration
	if ConfigListenAndServeWait > 0 {
		configListenAndServeWait = ConfigListenAndServeWait
	} else {
		configListenAndServeWait = DefaultListenAndServeWait
	}

	//create router
	r.router = chi.NewRouter()
	//Add middleware if have any
	if len(r.middlewares) > 0 {
		r.router.Use(r.middlewares...)
	}
	//create rest server
	r.restServer = &http.Server{
		Addr:      addr + ":" + port,
		Handler:   r.router,
		TLSConfig: r.tlsConfig,
	}
	//build the routes into the router
	r.buildRoutes(routes)
	//build the handles into the router
	r.buildHandles(handles)
	//launch the server
	r.LaunchServer()
	//wait for the server to start running successfully
	<-time.After(configListenAndServeWait)
	//set started to true
	r.started = true

	return
}

//Stop will attempt to shutdown the rest server, set its internal
// start flag to false and set all internal pointers created at start
// to nil
func (r *Router) Stop() (err error) {
	r.Lock()
	defer r.Unlock()

	var configServerShutdownTimeout time.Duration

	//check if the rest server is actually started
	if !r.started {
		err = errors.New(ErrRouterNotStarted)
		return
	}
	//validate and set configuration
	if ConfigServerShutdownTimeout > 0 {
		configServerShutdownTimeout = ConfigListenAndServeWait
	} else {
		configServerShutdownTimeout = DefaultServerShutdownTimeout
	}
	//create context with cancel
	ctx, cancel := context.WithTimeout(context.Background(), configServerShutdownTimeout)
	defer cancel()
	//shutdown the rest server
	if err = r.restServer.Shutdown(ctx); err != nil {
		return
	}
	//wait for the server to return with r.Done()
	r.Wait()
	//range through routes and delete
	for key := range r.routes {
		//delete from map
		delete(r.routes, key)
	}
	//remove rest server and router
	r.restServer, r.router = nil, nil
	//set started to false to signify that router has been stopped
	r.started = false

	return
}

//Close can be used to clean up any stored pointers that are created on new
// use this function to prepare to set the router pointer to nil, keep in mind
// that once this function is successful, the pointer can't be re-used
func (r *Router) Close() (err error) {
	r.Lock()
	defer r.Unlock()

	//only unset internal pointers if not started
	if r.started {
		err = errors.New(ErrRouterStarted)

		return
	}
	//close internal pointers
	//set internal configuration to defaults
	r.started = false
	//set internal pointers to nil
	r.router, r.restServer, r.routes, r.handles = nil, nil, nil, nil
	r.middlewares = nil
	return
}

//LaunchServer will create an anonymous goRoutine that will block while listening and
// serving and received requests
func (r *Router) LaunchServer() {
	//launch the rest server go routine which blocks until the server closes
	r.Add(1)
	go func() {
		defer r.Done()

		//we print the error for the simple fact that this is a single ended goRoutine that
		// isn't meant to return anything, we could also ignore the error by adding the following
		// with the commented out code below:
		// if err := r.restServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		if r.usingSecurity {
			if err := r.restServer.ListenAndServeTLS(r.certfile, r.keyfile); err != nil {
				r.logln(err.Error())
			}
		} else {
			if err := r.restServer.ListenAndServe(); err != nil {
				r.logln(err.Error())
			}
		}
	}()
}

//buildRoutes creates routes for the REST server
func (r *Router) buildRoutes(routes []RouteConfiguration) {
	//strong assumption that routes have already been validated
	//add all the routes to the internal map
	for _, route := range routes {
		r.routes[route.Route+route.Method] = route
	}
	//add all of the configurations to the router
	for _, route := range routes {
		r.router.MethodFunc(route.Method, route.Route, route.HandleFx)
	}
}

//buildRoutes creates routes for the REST server
func (r *Router) buildHandles(handles []HandleConfiguration) {
	//strong assumption that routes have already been validated
	//add all the routes to the internal map
	for _, handle := range handles {
		r.handles[handle.Route] = handle
	}
	//add all of the configurations to the router
	for _, handle := range handles {
		if !handle.Pattern {
			r.router.Handle(handle.Route, handle.HandleFx)
		} else {
			r.router.Mount(handle.Route, handle.HandleFx)
		}
	}
}

//logln can be used to write to the log provided at startup
func (r *Router) logln(info string) {
	if r.log != nil {
		r.log.Println(LogPrefix + info)
	}
}
