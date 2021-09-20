package search

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/n-creativesystem/saml-proxy/internal/options"
	"github.com/spf13/cast"
)

type MapSearch struct {
	keyDelim string
	value    map[string]interface{}
}

func New(value map[string]interface{}) MapSearch {
	return MapSearch{
		value:    value,
		keyDelim: ".",
	}
}

func (v *MapSearch) GetSliceString(lcaseKey string) ([]string, error) {
	path := strings.Split(lcaseKey, v.keyDelim)
	val := v.searchIndexableWithPathPrefixes(v.value, path)
	if val == nil {
		return nil, fmt.Errorf("no data found: %s", lcaseKey)
	}
	if v, err := cast.ToStringE(val); err == nil {
		return []string{v}, nil
	}
	if v, err := cast.ToStringSliceE(val); err == nil {
		return v, nil
	}
	return nil, fmt.Errorf("no support type %T", val)
}

func (v *MapSearch) GetString(lcaseKey string) (string, error) {
	val, err := v.GetSliceString(lcaseKey)
	if err != nil {
		return "", err
	}
	if len(val) == 0 {
		return "", fmt.Errorf("no data found: %s", lcaseKey)
	}
	return val[0], nil
}

func (v *MapSearch) searchIndexableWithPathPrefixes(source interface{}, path []string) interface{} {
	if len(path) == 0 {
		return source
	}

	// search for path prefixes, starting from the longest one
	for i := len(path); i > 0; i-- {
		prefixKey := strings.ToLower(strings.Join(path[0:i], v.keyDelim))

		var val interface{}
		switch sourceIndexable := source.(type) {
		case []interface{}:
			val = v.searchSliceWithPathPrefixes(sourceIndexable, prefixKey, i, path)
		case map[string]interface{}:
			val = v.searchMapWithPathPrefixes(sourceIndexable, prefixKey, i, path)
		}
		if val != nil {
			return val
		}
	}

	// not found
	return nil
}

func (v *MapSearch) searchSliceWithPathPrefixes(sourceSlice []interface{}, prefixKey string, pathIndex int, path []string) interface{} {
	// if the prefixKey is not a number or it is out of bounds of the slice
	index, err := strconv.Atoi(prefixKey)
	if err != nil || len(sourceSlice) <= index {
		return nil
	}

	next := sourceSlice[index]

	// Fast path
	if pathIndex == len(path) {
		return next
	}

	switch n := next.(type) {
	case map[interface{}]interface{}:
		return v.searchIndexableWithPathPrefixes(cast.ToStringMap(n), path[pathIndex:])
	case map[string]interface{}, []interface{}:
		return v.searchIndexableWithPathPrefixes(n, path[pathIndex:])
	default:
		// got a value but nested key expected, do nothing and look for next prefix
	}

	// not found
	return nil
}

func (v *MapSearch) searchMapWithPathPrefixes(sourceMap map[string]interface{}, prefixKey string, pathIndex int, path []string) interface{} {
	next, ok := sourceMap[prefixKey]
	if !ok {
		return nil
	}

	// Fast path
	if pathIndex == len(path) {
		return next
	}

	// Nested case
	switch n := next.(type) {
	case map[interface{}]interface{}:
		return v.searchIndexableWithPathPrefixes(cast.ToStringMap(n), path[pathIndex:])
	case map[string]interface{}, []interface{}:
		return v.searchIndexableWithPathPrefixes(n, path[pathIndex:])
	default:
		// got a value but nested key expected, do nothing and look for next prefix
	}

	// not found
	return nil
}

func (v *MapSearch) InjectHeader(header []options.Header, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		wHeader := rw.Header()
		for _, header := range header {
			for _, value := range header.Values {
				if values, err := v.GetSliceString(value.ClaimName); err == nil {
					for _, v := range values {
						wHeader.Add(header.Name, value.Prefix+v)
					}
				}
			}
		}
		next.ServeHTTP(rw, r)
	})
}
