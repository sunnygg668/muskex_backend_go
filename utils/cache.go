package utils

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/golang-lru"
	"hash/fnv"
	"log"
	"sync"
	"time"
)

var lcache *lru.Cache

func init() {
	var err error
	lcache, err = lru.New(80000)
	if err != nil {
		log.Println("lru cache error", err)
	} else {
		log.Println("lru cache init")
	}
}

func CacheFromLruWithFixKey(key string, myfunc func() (interface{}, error)) (interface{}, error) {
	res, ok := lcache.Get(key)
	if ok {
		log.Println("lru hit key", key)
		return res, nil
	}
	log.Println("lru miss key", key)
	obj, err := gp.Do(key, myfunc)
	if err != nil {
		log.Println("lru Group.Do load err", key, "err", err)
		return nil, err
	}
	_, ok = lcache.Get(key)
	if !ok {
		lcache.Add(key, obj)
	}
	return obj, err
}
func CacheFromLru(version int, key string, ttl int, myfunc func() (interface{}, error)) (interface{}, error) {
	if ttl > 0 {
		expire, _ := CalcExpirationPercentPadFromNow(int64(ttl), key)
		key = fmt.Sprintf("%s-%d-%d", key, version, expire)
	}
	return CacheFromLruWithFixKey(key, myfunc)

}
func CalcExpiration(ttl int64, key string) int64 {
	if ttl < 60 {
		log.Println("错误的ttl 应该大于60")
	}
	//we calculate the non discrete expiration, relative to current time
	now := time.Now().Unix()
	expires := now
	var padding int64 = 0
	h := fnv.New32a()
	h.Write([]byte(key))
	padding = int64(h.Sum32()) % 60
	//ran := rand.New(rand.NewSource(padding))
	//rvalue := ran.Int63n(60)
	ttl -= padding

	expires += (ttl - (expires % ttl))
	//log.Println("padding:=", padding, "ttl",ttl,"expires",expires,"now",now,"e-n",expires-now)
	return expires
}

func CalcExpirationPercentPadFromNow(ttlSec int64, key string) (expires, padding int64) {
	return CalcExpirationPercentPadFromTime(time.Now().UnixMilli(), ttlSec, key)
}
func CalcExpirationPercentPadFromTime(timeMsec int64, ttlSec int64, key string) (expires, padding int64) {
	if ttlSec < 1 {
		log.Println("错误的ttl 应该大于1")
	}
	ttlMilli := ttlSec * 1000
	//we calculate the non discrete expiration, relative to current time
	//now:=time.Now().UnixMilli()
	var randCut int64 = 0
	h := fnv.New32a()
	h.Write([]byte(key))
	randCut = int64(h.Sum32()) % (ttlMilli * 10 / 100)
	//ran := rand.New(rand.NewSource(padding))
	//rvalue := ran.Int63n(60)
	ttlMilli += randCut
	padding = (ttlMilli - (timeMsec % ttlMilli))
	expires = timeMsec + padding
	//log.Println(time.Unix(timeMsec/1000, timeMsec%1000*1e6), "rawTTL", ttlSec, "padding:=", padding, "ttl", ttlMilli, "expires", expires, "now", timeMsec, "e-n", expires-timeMsec)
	return
}

// call is an in-flight or completed Do call
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

var gp = group{}

// Group represents a class of work and forms a namespace in which
// units of work can be executed with duplicate suppression.
type group struct {
	mu sync.Mutex       // protects m
	m  map[string]*call // lazily initialized
}

// Do executes and returns the results of the given function, making
// sure that only one execution is in-flight for a given key at a
// time. If a duplicate comes in, the duplicate caller waits for the
// original to complete and receives the same results.
func (g *group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}

var addr = "13.229.187.76:56379"
var password = "YfJJExSdM"
var Rdb = redis.NewClient(&redis.Options{
	Addr:     addr,
	Password: password,
	DB:       10,
})
