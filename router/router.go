package router

import (
	"context"
	"net/http"
	"regexp"
)

type contextKey int

const (
	varsKey contextKey = iota
)

type Route struct {
	Path        string
	Method      string
	handlerFunc http.HandlerFunc
	isRegex     bool
}

type Router struct {
	routes []Route
}

func (rtr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range rtr.routes {
		isMatch, params := route.Match(r)
		if !isMatch {
			continue
		}

		ctx := context.WithValue(r.Context(), varsKey, params)
		route.handlerFunc.ServeHTTP(w, r.WithContext(ctx))

		return
	}

	http.NotFound(w, r)
}

func (rtr *Router) Route(method, path string, handlerFunc http.HandlerFunc, isRegex bool) {
	route := Route{
		Method:      method,
		Path:        path,
		handlerFunc: handlerFunc,
		isRegex:     isRegex,
	}
	rtr.routes = append(rtr.routes, route)
}

func (rtr *Route) Match(r *http.Request) (bool, map[string]string) {
	if !rtr.isRegex {
		if r.Method == rtr.Method && r.URL.Path == rtr.Path {
			return true, nil
		}

		return false, nil
	}

	regex := regexp.MustCompile(rtr.Path)

	match := regex.FindStringSubmatch(r.URL.Path)
	if match == nil {
		return false, nil
	}

	params := make(map[string]string)
	groupNames := regex.SubexpNames()
	for i, group := range match {
		params[groupNames[i]] = group
	}

	return true, params
}

// Returns the current request params, if any.
func RequestParams(r *http.Request) map[string]string {
	if value := r.Context().Value(varsKey); value != nil {
		return value.(map[string]string)
	}

	return nil
}
