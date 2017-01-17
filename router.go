package rest

import (
	"regexp"
	"strings"
)

type Routable interface {
	Get(string) *Route
	Match(string) *Route
	SetParamNames(*Route, []string)
	SetParamValues(*Route, []interface{})
}

type router struct {
	prefix     string
	Routes     map[string]*Route
	routingMap map[*Route]*regexp.Regexp
}

type routeBuilder struct {
	path      string
	actions   map[string]Action
	routeName string
}

var routers map[string]router

func init() {
	routers = make(map[string]router)
	routers["default"] = router{"", make(map[string]*Route), make(map[*Route]*regexp.Regexp)}
}

func NewRouter(name string) router {
	routers[name] = router{"", make(map[string]*Route), make(map[*Route]*regexp.Regexp)}
	return routers[name]
}

func DefaultRouter() router {
	return routers["default"]
}

func Router(name string) router {
	return routers[name]
}

func (r router) Prefix(prefix string) router {
	r.prefix = prefix
	return r
}

func NewRoute() *routeBuilder {
	return &routeBuilder{actions: make(map[string]Action)}
}

func (rb *routeBuilder) Named(name string) *routeBuilder {
	rb.routeName = name
	return rb
}

func (rb *routeBuilder) For(path string) *routeBuilder {
	rb.path = path
	return rb
}

func (rb *routeBuilder) With(method string, action Action) *routeBuilder {
	rb.actions[method] = action
	return rb
}

func (rb *routeBuilder) And(method string, action Action) *routeBuilder {
	return rb.With(method, action)
}

func (ro router) RouteMap(rbs ...*routeBuilder) router {
	for _, routeBuilder := range rbs {
		route := &Route{
			Path:    ro.prefix + routeBuilder.path,
			actions: routeBuilder.actions,
		}

		ro.initRoute(route)

		if routeBuilder.routeName == "" {
			ro.Routes[routeBuilder.path] = route
		}
		ro.Routes[routeBuilder.routeName] = route
	}

	return ro
}

func (ro router) initRoute(route *Route) {
	toSub := regParam.FindAllStringSubmatch(route.Path, -1)

	regString := route.Path

	if len(toSub) > 0 {
		params := make([]string, len(toSub))

		for i, v := range toSub {
			whole, param, pType, regex := v[0], v[1], v[2], `([^/]+)`

			params[i] = param

			if len(pType) > 1 {
				if r, ok := regMap[pType[1:]]; ok {
					regex = r
				}
			}
			regString = strings.Replace(regString, whole, regex, -1)
		}
		ro.SetParamNames(route, params)
	}

	ro.routingMap[route] = regexp.MustCompile(regString + "/?")
}

func (ro router) Get(name string) *Route {
	return ro.Routes[name]
}

func (ro router) Match(test string) *Route {
	for route, regex := range ro.routingMap {
		matches := regex.FindStringSubmatch(test)
		if matches != nil && matches[0] == test {
			values := make([]interface{}, len(matches[1:]))

			for i, m := range matches[1:] {
				values[i] = m
			}
			ro.SetParamValues(route, values)

			return route
		}
	}

	return nil
}

func (ro router) SetParamNames(r *Route, pn []string) {
	r.ParamNames = pn
}

func (ro router) SetParamValues(r *Route, pv []interface{}) {
	r.ParamValues = pv
}
