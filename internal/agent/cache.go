package agent

type Cache interface {
	Add(key string, value interface{}) error
	Get(key string) (interface{}, error)
	Delete(key string) error
}

type InMemoryCache struct {
}

func (c *InMemoryCache) Add(key string, value interface{}) error {
	return nil
}

func (c *InMemoryCache) Get(key string) (interface{}, error) {
	return nil, nil
}

func (c *InMemoryCache) Delete(key string) error {
	return nil
}
