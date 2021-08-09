package render

import pkgJson "encoding/json"

var JSON = &json{}

type json struct{}

var _ Render = (*json)(nil)

func (j *json) Marshal(obj interface{}) ([]byte, error) {
	return pkgJson.Marshal(obj)
}

func (j *json) Unmarshal(buf []byte, obj interface{}) error {
	return pkgJson.Unmarshal(buf, obj)
}
