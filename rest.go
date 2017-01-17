package rest

import "net/http"

const (
	MethodGET    = "GET"
	MethodPOST   = "POST"
	MethodPUT    = "PUT"
	MethodDELETE = "DELETE"
)

type ResponseWriter struct {
	http.ResponseWriter
	Size       int
	StatusCode int
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)

	if err == nil {
		rw.Size += size
	}

	return size, err
}

func (rw *ResponseWriter) WriteHeader(i int) {
	rw.StatusCode = i

	rw.ResponseWriter.WriteHeader(i)
}

func (rw *ResponseWriter) Written() bool {
	return rw.Size > 0 && rw.StatusCode > 0
}

type Request struct {
	*http.Request
	Data requestData
}

type requestData struct {
	data map[interface{}]interface{}
}

func (rd requestData) Set(key, value interface{}) {
	rd.data[key] = value
}

func (rd requestData) Get(key interface{}) interface{} {
	if value, ok := rd.data[key]; ok {
		return value
	}

	return nil
}

// Action is a type for all controller actions
type Action func(Context) ResponseSender

// Interceptor is a type for adding an intercepting the request before it is processed
type Interceptor func(Context) bool

// Middleware is a type for adding middleware for the request
type Middleware func(Context)

// Handler implements http.Handler and contains the router and controllers for the REST api
type handler struct {
	router       Routable
	interceptors []Interceptor
	middlewares  []Middleware
}

// NewHandler returns a new Handler with router initialized
func NewHandler(r Routable) *handler {
	return &handler{r, make([]Interceptor, 0), make([]Middleware, 0)}
}

func (h *handler) Intercept(i Interceptor) {
	h.interceptors = append(h.interceptors, i)
}

func (h *handler) Use(m Middleware) {
	h.middlewares = append(h.middlewares, m)
}

func (h *handler) invokeInterceptors(c Context) bool {
	result := true
	for i := 0; i < len(h.interceptors) && result; i++ {
		result = h.interceptors[i](c)
	}

	return result
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route := h.router.Match(r.URL.Path)
	if route == nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	action, actionExists := route.actions[r.Method]
	if !actionExists {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	context := Context{
		Request:        Request{r, requestData{make(map[interface{}]interface{})}},
		ResponseWriter: &ResponseWriter{w, 0, 0},
		Route:          route,
		middlewares:    h.middlewares,
		action:         action,
	}

	if ok := h.invokeInterceptors(context); !ok {
		// maybe check to see if response and header/status has been written
		// if not, then probably should do something
		return
	}

	context.run()

	return
}
