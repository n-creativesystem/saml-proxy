package utils

import (
	"github.com/spf13/cast"
)

func MapFlatten(src map[string]interface{}, dest map[string][]string) {
	for key, value := range src {
		switch v := value.(type) {
		case map[string]interface{}:
			MapFlatten(v, dest)
		case []interface{}:
			if len(v) > 1 {
				dest[key] = make([]string, 0, len(v))
				for i := 0; i < len(v); i++ {
					dest[key] = append(dest[key], cast.ToString(v[i]))
				}
			} else {
				dest[key] = []string{cast.ToString(v[0])}
			}
		default:
			dest[key] = []string{cast.ToString(v)}
		}
	}
}
