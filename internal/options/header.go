package options

type Header struct {
	Name   string        `json:"name"`
	Values []HeaderValue `json:"values"`
}

type Headers []Header

func (headers Headers) Clone() Headers {
	results := make(Headers, 0, len(headers))
	for _, h := range headers {
		header := Header{
			Name:   h.Name,
			Values: HeaderValues(h.Values).Clone(),
		}
		results = append(results, header)
	}
	return results
}

type HeaderValue struct {
	*ClaimSource `json:",omitempty"`
}

func (h *HeaderValue) Clone() HeaderValue {
	c := h.ClaimSource.Clone()
	return HeaderValue{
		ClaimSource: &c,
	}
}

type HeaderValues []HeaderValue

func (values HeaderValues) Clone() HeaderValues {
	results := make(HeaderValues, 0, len(values))
	for _, v := range values {
		results = append(results, v.Clone())
	}
	return results
}

type ClaimSource struct {
	ClaimName         string `json:"claim"`
	Prefix            string `json:"prefix"`
	BasicAuthPassword string `json:"basicAuthPassword"`
}

func (c *ClaimSource) Clone() ClaimSource {
	return ClaimSource{
		ClaimName:         c.ClaimName,
		Prefix:            c.Prefix,
		BasicAuthPassword: c.BasicAuthPassword,
	}
}
