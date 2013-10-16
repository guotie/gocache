package gocache

import (
	"encoding/json"
	"errors"
	"github.com/bradfitz/gomemcache/memcache"
	"log"
	"strings"
	"time"
)

// Wraps the Memcached client to meet the Cache interface.
type MemcachedCache struct {
	*memcache.Client
	defaultExpiration time.Duration
}

func NewMemcachedCache(hostList []string, defaultExpiration time.Duration) MemcachedCache {
	return MemcachedCache{memcache.New(hostList...), defaultExpiration}
}

func (c MemcachedCache) Open(config string) (Cache, error) {
	var cf map[string]string
	json.Unmarshal([]byte(config), &cf)
	hosts, ok := cf["hosts"]
	if !ok {
		return nil, errors.New("config invalid: no hosts.")
	}
	hostList := strings.Split(hosts, ",")
	for i, host := range hostList {
		hostList[i] = strings.TrimSpace(host)
	}

	return NewMemcachedCache(hostList, 0), nil
}

func (c MemcachedCache) Set(key string, value interface{}, expires time.Duration) error {
	return c.invoke((*memcache.Client).Set, key, value, expires)
}

func (c MemcachedCache) Add(key string, value interface{}, expires time.Duration) error {
	return c.invoke((*memcache.Client).Add, key, value, expires)
}

func (c MemcachedCache) Replace(key string, value interface{}, expires time.Duration) error {
	//return c.invoke((*memcache.Client).Replace, key, value, expires)
	return errors.New("Replace is not support currently.")
}

func (c MemcachedCache) Get(key string, ptrValue interface{}) error {
	item, err := c.Client.Get(key)
	if err != nil {
		return convertMemcacheError(err)
	}
	return Deserialize(item.Value, ptrValue)
}

func (c MemcachedCache) GetMulti(keys ...string) (Getter, error) {
	items, err := c.Client.GetMulti(keys)
	if err != nil {
		return nil, convertMemcacheError(err)
	}
	return ItemMapGetter(items), nil
}

func (c MemcachedCache) Delete(key string) error {
	return convertMemcacheError(c.Client.Delete(key))
}

func (c MemcachedCache) Increment(key string, delta uint64) (newValue uint64, err error) {
	newValue, err = c.Client.Increment(key, delta)
	return newValue, convertMemcacheError(err)
}

func (c MemcachedCache) Decrement(key string, delta uint64) (newValue uint64, err error) {
	newValue, err = c.Client.Decrement(key, delta)
	return newValue, convertMemcacheError(err)
}

func (c MemcachedCache) Flush() error {
	err := errors.New("gocache: can not flush memcached.")
	log.Println(err)
	return err
}

func (c MemcachedCache) invoke(f func(*memcache.Client, *memcache.Item) error,
	key string, value interface{}, expires time.Duration) error {

	switch expires {
	case DEFAULT:
		expires = c.defaultExpiration
	case FOREVER:
		expires = time.Duration(0)
	}

	b, err := Serialize(value)
	if err != nil {
		return err
	}
	return convertMemcacheError(f(c.Client, &memcache.Item{
		Key:        key,
		Value:      b,
		Expiration: int32(expires / time.Second),
	}))
}

// Implement a Getter on top of the returned item map.
type ItemMapGetter map[string]*memcache.Item

func (g ItemMapGetter) Get(key string, ptrValue interface{}) error {
	item, ok := g[key]
	if !ok {
		return ErrCacheMiss
	}

	return Deserialize(item.Value, ptrValue)
}

func convertMemcacheError(err error) error {
	switch err {
	case nil:
		return nil
	case memcache.ErrCacheMiss:
		return ErrCacheMiss
	case memcache.ErrNotStored:
		return ErrNotStored
	}

	log.Println("gocache:", err)
	return err
}
