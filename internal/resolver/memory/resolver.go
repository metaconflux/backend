package memory

import "github.com/metaconflux/backend/internal/resolver"

type Resolver struct {
	resolver.IResolver
	data map[string]string
}

func NewResolver() resolver.IResolver {
	return Resolver{
		data: make(map[string]string),
	}
}

func (r Resolver) Get(key string) (string, error) {
	result, ok := r.data[key]
	if !ok {
		return result, resolver.ErrNotFound
	}

	return result, nil
}

func (r Resolver) Set(key string, val string) error {
	r.data[key] = val
	return nil
}
