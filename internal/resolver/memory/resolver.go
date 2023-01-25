package memory

import (
	"time"

	"github.com/metaconflux/backend/internal/resolver"
)

type Entry struct {
	Value   string
	Timeout time.Time
}

type Resolver struct {
	resolver.IResolver
	data map[string]Entry
}

func NewResolver() resolver.IResolver {
	return Resolver{
		data: make(map[string]Entry),
	}
}

func (r Resolver) Get(key string) (string, error) {
	result, ok := r.data[key]
	if !ok {
		return result.Value, resolver.ErrNotFound
	}

	if result.Timeout.Unix() > 0 && result.Timeout.Before(time.Now()) {
		return result.Value, resolver.ErrNotFound
	}

	return result.Value, nil
}

func (r Resolver) Set(key string, val string, timeout int64) error {
	var t time.Time

	if timeout > 0 {
		t = time.Now().Add(time.Duration(timeout) * time.Minute)
	}
	entry := Entry{
		Value:   val,
		Timeout: t,
	}
	r.data[key] = entry
	return nil
}
