package saml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUrlResources(t *testing.T) {
	values := []struct {
		Name    string
		URL     string
		Methods []string
		Roles   []string
		Fn      func(resource *BaseResource) func(t *testing.T)
	}{
		{
			Name:    "wildcard",
			URL:     "/*",
			Methods: []string{"*"},
			Roles:   []string{"role1", "role2"},
			Fn: func(resource *BaseResource) func(t *testing.T) {
				return func(t *testing.T) {
					assert.NotNil(t, resource)
					resources := BaseResources{*resource}.ToResources()
					assert.True(t, resources.LookUp("get", "/", "role1"))
					assert.True(t, resources.LookUp("get", "/api", "role1"))
					assert.True(t, resources.LookUp("get", "/api/v1", "role1"))
				}
			},
		},
		{
			Name:    "default",
			URL:     "/api/v1/*",
			Methods: []string{"get", "post", "put", "delete"},
			Roles:   []string{"role1", "role2"},
			Fn: func(resource *BaseResource) func(t *testing.T) {
				return func(t *testing.T) {
					assert.NotNil(t, resource)
					resources := BaseResources{*resource}.ToResources()
					assert.True(t, resources.LookUp("get", "/api/v1/test", "role1"))
				}
			},
		},
		{
			Name:    "all",
			URL:     "/api/v1/test",
			Methods: []string{"*"},
			Roles:   []string{"role1", "role2"},
			Fn: func(resource *BaseResource) func(t *testing.T) {
				return func(t *testing.T) {
					assert.NotNil(t, resource)
				}
			},
		},
		{
			Name:    "method nothing",
			URL:     "/api/v1/test",
			Methods: []string{},
			Roles:   []string{"role1", "role2"},
			Fn: func(resource *BaseResource) func(t *testing.T) {
				return func(t *testing.T) {
					assert.Nil(t, resource)
				}
			},
		},
		{
			Name:    "role nothing",
			URL:     "/api/v1/test",
			Methods: []string{"*"},
			Roles:   []string{},
			Fn: func(resource *BaseResource) func(t *testing.T) {
				return func(t *testing.T) {
					assert.Nil(t, resource)
				}
			},
		},
		{
			Name:    "url nothing",
			URL:     "",
			Methods: []string{"*"},
			Roles:   []string{"role1", "role2"},
			Fn: func(resource *BaseResource) func(t *testing.T) {
				return func(t *testing.T) {
					assert.Nil(t, resource)
				}
			},
		},
	}
	for _, value := range values {
		if value.Fn != nil {
			v := NewBaseResource(value.URL, value.Methods, value.Roles)
			t.Run(value.Name, value.Fn(v))
		}
	}
}
