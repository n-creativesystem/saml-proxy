package options

type Upstream struct {
	URL                   string     `json:"url"`
	Mapping               Mapping    `json:"mapping"`
	Resources             []Resource `json:"resources"`
	InSecureSkipTlsVerify bool       `json:"insecureSkipTlsVerify"`
	InjectRequestHeaders  []Header   `json:"injectRequestHeaders"`
}
type Mapping struct {
	Name     string `json:"name"`
	Multiple bool   `json:"multiple"`
}
