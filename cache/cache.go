// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package cache provides a simple in-memory caching mechanism with expiration support.
// It uses sync.Map for thread-safe operations and supports string key-value pairs.
package cache

import (
	"errors"
	"sync"
	"time"
)

// Cache defines the interface for cache operations.
type Cache interface {
	// SetString stores a string value in the cache with an expiration time.
	SetString(key string, value string, expiration int) error

	// GetString retrieves a string value from the cache.
	GetString(key string) (string, error)
}

// cache implements the Cache interface using sync.Map as the underlying storage.
type cache struct {
	buckets sync.Map
}

// item represents a cache entry with a value and expiration time.
type item struct {
	value      string
	expiration int64
}

var (
	// ErrKeyNotFound is returned when a key is not found in the cache.
	ErrKeyNotFound = errors.New("key not found")

	// ErrKeyExpired is returned when a key has expired in the cache.
	ErrKeyExpired = errors.New("key has expired")
)

// SetString stores a string value in the cache with the specified expiration time.
//
// Parameters:
//   - key: The unique identifier for the cache entry.
//   - value: The string value to be stored.
//   - expiration: The number of seconds after which the entry should expire.
//     If set to 0, the entry will not expire.
//
// Returns:
//   - error: Always returns nil in this implementation.
//
// Example:
//
//	c := cache.New()
//	err := c.SetString("user_1", "John Doe", 3600) // Expires in 1 hour
//	if err != nil {
//	    log.Printf("Failed to set cache: %v", err)
//	}
func (c *cache) SetString(key string, value string, expiration int) error {
	var exp int64
	if expiration > 0 {
		// Calculate the expiration time in nanoseconds
		exp = time.Now().Add(time.Duration(expiration) * time.Second).UnixNano()
	} else {
		// Use 0 to represent no expiration
		exp = 0
	}

	// Store the item in the sync.Map
	c.buckets.Store(key, item{
		value:      value,
		expiration: exp,
	})

	return nil
}

// GetString retrieves a string value from the cache.
//
// Parameters:
//   - key: The unique identifier of the cache entry to retrieve.
//
// Returns:
//   - string: The value associated with the key if found and not expired.
//   - error: ErrKeyNotFound if the key doesn't exist, ErrKeyExpired if the key has expired,
//     or another error if there's an issue with the cache item.
//
// Example:
//
//	c := cache.New()
//	c.SetString("user_1", "John Doe", 3600)
//	value, err := c.GetString("user_1")
//	if err != nil {
//	    if err == cache.ErrKeyNotFound {
//	        log.Println("Key not found in cache")
//	    } else if err == cache.ErrKeyExpired {
//	        log.Println("Cache entry has expired")
//	    } else {
//	        log.Printf("Error retrieving from cache: %v", err)
//	    }
//	} else {
//	    log.Printf("Retrieved value: %s", value)
//	}
func (c *cache) GetString(key string) (string, error) {
	// Attempt to load the v from the sync.Map
	val, ok := c.buckets.Load(key)
	if !ok {
		return "", ErrKeyNotFound
	}

	// Type assert the loaded value to an v
	v, ok := val.(item)
	if !ok {
		return "", errors.New("invalid cache v")
	}

	// Check if the v has expired
	if v.expiration != 0 && time.Now().UnixNano() > v.expiration {
		// Remove the expired v from the cache
		c.buckets.Delete(key)
		return "", ErrKeyExpired
	}

	// Return the value if it's valid and not expired
	return v.value, nil
}

// New creates and returns a new instance of Cache.
//
// Returns:
//   - Cache: A new Cache instance.
//
// Example:
//
//	c := cache.New()
//	// Now you can use c to set and get cache entries
func New() Cache {
	return &cache{}
}
