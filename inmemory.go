package gocache

import (
	"fmt"
	"github.com/pmylund/go-cache"
	"log"
	"reflect"
	"time"
)

type InMemoryCache struct {
	cache.Cache
}

func NewInMemoryCache(defaultExpiration time.Duration) InMemoryCache {
	return InMemoryCache{*cache.New(defaultExpiration, time.Minute)}
}

func (c InMemoryCache) Open(config string) (Cache, error) {
	return NewInMemoryCache(0), nil
}

func (c InMemoryCache) Get(key string, ptrValue interface{}) error {
	value, found := c.Cache.Get(key)
	if !found {
		return ErrCacheMiss
	}

	v := reflect.ValueOf(ptrValue)
	if v.Type().Kind() == reflect.Ptr && v.Elem().CanSet() {
		v.Elem().Set(reflect.ValueOf(value))
		return nil
	}

	err := fmt.Errorf("gocache: attempt to get %s, but can not set value %v", key, v)
	log.Println(err)
	return err
}

func (c InMemoryCache) GetMulti(keys ...string) (Getter, error) {
	return c, nil
}

func (c InMemoryCache) Set(key string, value interface{}, expires time.Duration) error {
	// NOTE: go-cache understands the values of DEFAULT and FOREVER
	c.Cache.Set(key, value, expires)
	return nil
}

func (c InMemoryCache) Add(key string, value interface{}, expires time.Duration) error {
	err := c.Cache.Add(key, value, expires)
	if err != nil {
		return ErrNotStored
	}
	return err
}

func (c InMemoryCache) Replace(key string, value interface{}, expires time.Duration) error {
	if err := c.Cache.Replace(key, value, expires); err != nil {
		return ErrNotStored
	}
	return nil
}

func (c InMemoryCache) Delete(key string) error {
	c.Cache.Delete(key)
	return nil
}

func (c InMemoryCache) Increment(key string, n uint64) (newValue uint64, err error) {
	newValue, err = c.Cache.IncrementUint64(key, n)
	if err != nil {
		return 0, ErrCacheMiss
	}
	return
}

func (c InMemoryCache) Decrement(key string, n uint64) (newValue uint64, err error) {
	newValue, err = c.Cache.DecrementUint64(key, n)
	if err != nil {
		return 0, ErrCacheMiss
	}
	return
}

func (c InMemoryCache) Flush() error {
	c.Cache.Flush()
	return nil
}
