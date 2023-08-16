package router

import "net/http"

type Route struct {
	Path        string
	Method      string
	handlerFunc http.HandlerFunc
}

type Router struct {
	routes []Route
}

func (rtr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range rtr.routes {
		if !route.Match(r) {
			continue
		}

		route.handlerFunc.ServeHTTP(w, r)
		return
	}

	http.NotFound(w, r)
}

func (rtr *Router) Route(method, path string, handlerFunc http.HandlerFunc) {
	route := Route{
		Method:      method,
		Path:        path,
		handlerFunc: handlerFunc,
	}
	rtr.routes = append(rtr.routes, route)
}

func (rtr *Route) Match(r *http.Request) bool {
	if r.Method != rtr.Method || r.URL.Path != rtr.Path {
		return false
	}

	return true
}
