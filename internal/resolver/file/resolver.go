package file

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/metaconflux/backend/internal/resolver"
)

type Entry struct {
	Value    string    `json:"value"`
	Lifetime time.Time `json:"lifetime"`
}

type ResolverMap map[string]Entry

type Resolver struct {
	resolver.IResolver
	store string
	lock  sync.Mutex
}

func NewResolver(file string) (resolver.IResolver, error) {
	dir := filepath.Dir(file)
	_, err := os.Stat(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.MkdirAll(dir, 0700)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	_, err = os.Stat(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			resolverMap := make(ResolverMap)
			data, err := json.Marshal(resolverMap)
			if err != nil {
				return nil, err
			}
			err = ioutil.WriteFile(file, data, 0700)
			if err != nil {
				return nil, err
			}
		}
	}
	return &Resolver{
		store: file,
	}, nil
}

func (r *Resolver) Get(key string) (string, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	data, err := ioutil.ReadFile(r.store)
	if err != nil {
		return "", err
	}
	var resolverMap ResolverMap
	err = json.Unmarshal(data, &resolverMap)
	if err != nil {
		return "", err
	}
	result, ok := resolverMap[key]
	if !ok {
		return result.Value, resolver.ErrNotFound
	}

	if result.Lifetime.Unix() > 0 && result.Lifetime.Before(time.Now()) {
		return result.Value, resolver.ErrNotFound
	}

	return result.Value, nil
}

func (r *Resolver) Set(key string, val string, timeout int64) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	data, err := ioutil.ReadFile(r.store)
	if err != nil {
		return err
	}
	var resolverMap ResolverMap
	err = json.Unmarshal(data, &resolverMap)
	if err != nil {
		return err
	}

	var t time.Time

	if timeout > 0 {
		t = time.Now().Add(time.Duration(timeout) * time.Minute)
	}
	entry := Entry{
		Value:    val,
		Lifetime: t,
	}

	resolverMap[key] = entry
	data, err = json.Marshal(resolverMap)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(r.store, data, 0700)
	if err != nil {
		return err
	}

	return nil
}
