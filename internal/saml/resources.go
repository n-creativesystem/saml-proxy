package saml

import (
	"net/http"
	"strings"

	"github.com/n-creativesystem/saml-proxy/typ"
	"github.com/ucarion/urlpath"
)

var allMethods = typ.StringSlice{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodConnect,
	http.MethodOptions,
	http.MethodTrace,
}

type Injection struct{}

type Matching interface {
	LookUp(method, url string, role string) bool
	MultipleLookUp(method, url string, roles []string) bool
}

type BaseResource struct {
	URL     string
	Methods []string
	Roles   []string
}

func NewBaseResource(url string, methods, roles []string) *BaseResource {
	if len(methods) == 0 {
		return nil
	}

	if len(roles) == 0 {
		return nil
	}

	if len(url) == 0 {
		return nil
	}
	valueMethods := make([]string, 0, len(allMethods))
	if len(methods) == 1 && methods[0] == "*" {
		valueMethods = append(valueMethods, allMethods.Copy()...)
	} else {
		for _, method := range methods {
			valueMethods = append(valueMethods, strings.ToUpper(method))
		}
	}
	valueRoles := make([]string, 0, len(roles))
	valueRoles = append(valueRoles, roles...)

	return &BaseResource{
		URL:     url,
		Methods: valueMethods,
		Roles:   valueRoles,
	}
}

type BaseResources []BaseResource

func (r BaseResources) ToResources() Matching {
	methods := make(urlPathResources)
	for _, resource := range r {
		sliceString := typ.StringSlice(resource.Roles)
		path := newPathRole(resource.URL, sliceString.ToMap())
		for _, method := range resource.Methods {
			method = strings.ToUpper(method)
			methods[method] = append(methods[method], path)
		}
	}
	return urlPathResources(methods)
}

type pathRole struct {
	path  urlpath.Path
	roles typ.IMap
}

func (p *pathRole) lookup(path string) bool {
	_, ok := p.path.Match(path)
	return ok
}

func (p *pathRole) exists(role string) bool {
	return p.roles.Exists(role)
}

func newPathRole(path string, roles typ.IMap) pathRole {
	return pathRole{
		path:  urlpath.New(path),
		roles: roles,
	}
}

type urlPathResources map[string][]pathRole

func (r urlPathResources) getPathRole(method, url string) (pathRole, bool) {
	method = strings.ToUpper(method)
	if resources, ok := r[method]; ok {
		for _, resource := range resources {
			ok := resource.lookup(url)
			if ok {
				return resource, true
			}
		}
	}
	return pathRole{}, false
}

func (r urlPathResources) LookUp(method, url string, role string) bool {
	resource, ok := r.getPathRole(method, url)
	if !ok {
		return false
	}
	return resource.exists(role)
}

func (r urlPathResources) MultipleLookUp(method, url string, roles []string) bool {
	resource, ok := r.getPathRole(method, url)
	if !ok {
		return false
	}
	for _, role := range roles {
		if resource.exists(role) {
			return true
		}
	}
	return false
}
