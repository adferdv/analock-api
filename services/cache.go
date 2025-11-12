package services

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/adfer-dev/analock-api/utils"
)

var cacheServiceInstance *cacheServiceImpl

type CacheService interface {
	CacheResource(f func() (interface{}, error), resource string, key string) (interface{}, error)
	EvictResourceItem(f CacheFunc, resource string, key string) (interface{}, error)
	EvictUserResource(resource string, userId int) error
}

type cacheServiceImpl struct {
	cache *cache
}

type CacheFunc func() (interface{}, error)

// Gets the singleton instance of the Cache Service
func GetCacheServiceInstance() *cacheServiceImpl {
	if cacheServiceInstance == nil {
		cacheServiceInstance = NewCacheService()
	}

	return cacheServiceInstance
}

// Builds a new Cache Service
func NewCacheService() *cacheServiceImpl {
	expirationTime, expirationParseErr := time.ParseDuration(os.Getenv("API_CACHE_EXPIRATION"))
	evictionInterval, intervalParseErr := time.ParseDuration(os.Getenv("API_CACHE_EVICTION_INTERVAL"))

	if expirationParseErr != nil {
		log.Fatalf(
			"Error when parsing cache exp from env variable: %s",
			expirationParseErr.Error(),
		)
	}

	if intervalParseErr != nil {
		log.Fatalf(
			"Error when parsing eviction interval from env variable: %s",
			intervalParseErr.Error(),
		)
	}
	return &cacheServiceImpl{cache: newCache(expirationTime, evictionInterval)}
}

// Caches the result of the given function or returns the already cached value if exists.
// When caching the resource, builds a key based on the concatenation of resource + key
func (cs *cacheServiceImpl) CacheResource(f func() (interface{}, error), resource string, key string) (interface{}, error) {
	fullKey := fmt.Sprintf("%s-%s", resource, key)
	cached, cacheErr := cs.cache.get(fullKey)

	if cacheErr == nil {
		log.Printf("CACHE HIT: key: %s, value: %+v\n", fullKey, cached)
		return cached, nil
	}

	fnRes, fnErr := f()

	if fnErr == nil {
		cs.cache.put(fullKey, fnRes)
	}

	return fnRes, fnErr
}

// Evicts all the cache entries whose keys starts with a concatenation of the given resource and user.
func (cs *cacheServiceImpl) EvictUserResource(resource string, userId uint) error {
	regex, regexErr := regexp.Compile(
		fmt.Sprintf("^%s-%s*", resource, utils.BuildUserCacheKey(userId)),
	)

	if regexErr != nil {
		utils.GetCustomLogger().Errorf(
			"Regex error on cache evict: %s",
			regexErr.Error(),
		)
		return regexErr
	}

	cs.cache.deleteIfMatches(regex)

	return nil
}

// Evicts the cache entry holding the key that results from the concatenation of resource + key params.
func (cs *cacheServiceImpl) EvictResourceItem(resource string, key string) {
	cs.cache.delete(fmt.Sprintf("%s-%s", resource, key))
}

type cache struct {
	entries        map[string]*cacheEntry
	evicter        *cacheEvicter
	mutex          sync.Mutex
	expirationTime time.Duration
}

type cacheEntry struct {
	entry interface{}
	time  time.Time
}

// Adds a new entry to the cache having the given key and value.
func (cache *cache) put(key string, value interface{}) {
	log.Printf("CACHE PUT: key: %s, value: %+v\n", key, value)
	cache.mutex.Lock()
	cache.entries[key] = &cacheEntry{entry: value, time: time.Now()}
	cache.mutex.Unlock()
}

// Gets the value of the entry with the given key.
// Returns error if no entry with that key was found.
func (cache *cache) get(key string) (*interface{}, error) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	result, present := cache.entries[key]

	if !present {
		return nil, errors.New("cache entry is not present")
	}

	return &result.entry, nil
}

// Deletes the entry that matches the given key from the cache.
func (cache *cache) delete(key string) {
	log.Printf("DELETE FROM CACHE: key: %s\n", key)
	cache.mutex.Lock()
	delete(cache.entries, key)
	cache.mutex.Unlock()
}

// Deletes the entries whose keys match the given regex pattern.
func (cache *cache) deleteIfMatches(regex *regexp.Regexp) {
	log.Printf("DELETE FROM CACHE: regex: %s\n", regex.String())
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	for key := range cache.entries {
		if regex.MatchString(key) {
			delete(cache.entries, key)
		}
	}
}

// Handles the eviction of expired cache entries.
func (cache *cache) handleEviction(currentTime time.Time) {
	utils.GetCustomLogger().Info("Running cache eviction...")
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	for key, value := range cache.entries {
		if currentTime.After(value.time.Add(cache.expirationTime)) {
			utils.GetCustomLogger().Infof("Evicting %s\n", key)
			delete(cache.entries, key)
		}
	}
}

// Builds a new cache and runs the eviction thread
func newCache(expirationTime time.Duration, evictionInterval time.Duration) *cache {
	cache := &cache{}
	evicter := &cacheEvicter{exitChannel: make(chan int), evictionInterval: evictionInterval}
	cache.entries = make(map[string]*cacheEntry)
	cache.expirationTime = expirationTime
	cache.evicter = evicter
	go cache.evicter.Run(cache)

	return cache
}

type cacheEvicter struct {
	evictionInterval time.Duration
	exitChannel      chan int
}

// Runs cache eviction each time determined by the evicter's eviction interval.
func (evicter *cacheEvicter) Run(c *cache) {
	ticker := time.NewTicker(evicter.evictionInterval)

	for {
		select {
		case currentTime := <-ticker.C:
			c.handleEviction(currentTime)
		case <-evicter.exitChannel:
			return
		}
	}
}

// Writes to the evicter's exit channel, ending the eviction worker.
func (evicter *cacheEvicter) Stop() {
	evicter.exitChannel <- 0
}
