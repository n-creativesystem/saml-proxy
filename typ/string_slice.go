package typ

type IMap interface {
	Exists(key string) bool
}

type mapString map[string]struct{}

func (m mapString) Exists(key string) bool {
	// return true of no key
	if len(m) == 0 {
		return true
	}
	_, ok := m[key]
	return ok
}

type StringSlice []string

func (slice StringSlice) Copy() []string {
	results := make([]string, 0, len(slice))
	for _, s := range slice {
		results = append(results, s)
	}
	return results
}

func (slice StringSlice) ToMap() IMap {
	results := make(mapString, len(slice))
	for _, s := range slice {
		results[s] = struct{}{}
	}
	return results
}
