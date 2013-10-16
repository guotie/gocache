gocache
=======

go cache

code copy from revel/cache


Usage:

   import "github.com/guotie/gocache"

1. Open
err := gocache.Open(driver)

driver can be "memory", "memcache"

2. use it

gocache.Get()
gocache.Set()
...