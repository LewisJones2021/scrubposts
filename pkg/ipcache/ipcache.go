package ipcache

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/lewisjones2021/scrubposts/pkg/background"
)

// cacheExpirationTime represents the duration, in minutes,
// for which an item is considered valid in the IP cache.
const cacheExpirationTime = 1

// IPCacheItem represents an item stored in the IP cache,
// containing information about the initial request time
// and the last seen time.
type IPCacheItem struct {
	InitialRequestTime time.Time
	LastSeenAT         time.Time
}

type Cache struct {
	cache map[string]*IPCacheItem
	mu    sync.RWMutex
}

func New() *Cache {
	c := &Cache{
		// Initialize the cache map.
		cache: make(map[string]*IPCacheItem),
		// Initialize a RWMutex for concurrent access control.
		mu: sync.RWMutex{},
	}

	// clear the ip cache every 5 minutes
	background.Go(BackgroundCacheClearer(c))

	return c
}

// backgroundCacheClearer clears up stale entries from the ipcache on a time.Ticker schedule.
func BackgroundCacheClearer(c *Cache) func() {
	// Return an anonymous function that clears expired cache items at a regular interval.
	return func() {
		// Execute the following code block repeatedly at intervals of 30 seconds.
		for range time.Tick(30 * time.Second) {
			// Anonymous function to safely acquire and release the mutex lock.
			func() {
				// Lock the read-write mutex to prevent concurrent access to the cache map.
				c.mu.Lock()
				// Ensure the mutex is released (unlocked) when the function exits.
				defer c.mu.Unlock()

				slog.Info("Checking IP cache for stale IP's")
				// Every 30 seconds, check every ip in the cache to see if it can be deleted.
				// Iterate over each IP address and its corresponding cache item in the cache.
				for ip, item := range c.cache {
					// Check if the initial request time for the IP address is before the current time minus the cache expiration time.
					// If the ip was cached over 5 minutes ago, go ahead and delete it.
					if item.InitialRequestTime.Before(time.Now().Add(-cacheExpirationTime * time.Minute)) {
						slog.Info("found ip to be deleted", "ip:", ip)
						fmt.Println(ip)
						// delete the ip from the cache
						delete(c.cache, ip)
					}
				}
			}()
		}
	}
}

// Set adds the item to the cache
func (c *Cache) Set(ip, id string) error {
	// Lock the mutex to ensure exclusive access to the cache while setting the item.
	c.mu.Lock()
	// Unlock the mutex when the function exits, regardless of whether it panics or returns normally.
	defer c.mu.Unlock()

	// Add the IP address and post ID combination to the cache.
	key := ip + "_" + id
	fmt.Println("Setting cache key:", key)
	c.cache[ip+"_"+id] = &IPCacheItem{
		// Set the initial request time to the current time.
		InitialRequestTime: time.Now(),
	}
	// no error
	return nil
}

// Has returns true if the IP address and post ID combination exists in the cache
func (c *Cache) Has(ip, id string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	key := ip + "_" + id
	_, ok := c.cache[ip+"_"+id]
	fmt.Println("Checking cache key:", key)
	return ok, nil
}
