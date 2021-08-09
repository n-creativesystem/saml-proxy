package render

type Render interface {
	Marshal(obj interface{}) ([]byte, error)
	Unmarshal(buf []byte, obj interface{}) error
}
