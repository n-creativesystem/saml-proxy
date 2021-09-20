package options

type Resource struct {
	Methods []string `json:"methods"`
	URL     string   `json:"url"`
	Roles   []string `json:"roles"`
}
