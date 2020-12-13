package cache

import "sync"

type Cache struct {
	values map[string]string
	mutex  sync.RWMutex
}

func New() Cache {
	return Cache{
		values: map[string]string{},
	}
}

func (c *Cache) Get(key string) string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	value, ok := c.values[key]
	if !ok {
		return ""
	}
	return value
}

func (c *Cache) Set(key string, value string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.values[key] = value
}

func (c *Cache) Remove(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.values, key)
}
