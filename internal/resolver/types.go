package resolver

import "fmt"

var ErrNotFound = fmt.Errorf("Not Found")

type IResolver interface {
	Get(key string) (string, error)
	Set(key string, val string) error
}
