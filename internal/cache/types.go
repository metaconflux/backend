package cache

type ICache interface {
	Push(object interface{}) (string, error)
	Get(id string, target interface{}) error
}
