package resolver

import "fmt"

var ErrNotFound = fmt.Errorf("Not Found")
var ErrLifetime = fmt.Errorf("Lifetime is over")

type IResolver interface {
	Get(key string) (string, error)
	Set(key string, val string, timeout int64) error
	Delete(key string) error
}
