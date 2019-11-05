package api

type Cache interface {
	Set(key string, data interface{}) error
	Get(key string) (interface{}, bool, error)
}