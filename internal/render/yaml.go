package render

import pkgYaml "gopkg.in/yaml.v3"

var YAML = &yaml{}

type yaml struct{}

func (y *yaml) Marshal(obj interface{}) ([]byte, error) {
	return pkgYaml.Marshal(obj)
}

func (y *yaml) Unmarshal(buf []byte, obj interface{}) error {
	return pkgYaml.Unmarshal(buf, obj)
}
