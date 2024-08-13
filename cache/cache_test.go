// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package cache

import (
	"errors"
	"testing"
	"time"
)

// TestCache runs unit tests for the Cache implementation.
func TestCache(t *testing.T) {
	// Create a new cache instance
	c := New()

	// Test SetString and GetString
	t.Run("Set and Get String", func(t *testing.T) {
		key := "test_key"
		value := "test_value"
		expiration := 5 // 5 seconds

		// Set the value
		err := c.SetString(key, value, expiration)
		if err != nil {
			t.Errorf("SetString failed: %v", err)
		}

		// Get the value immediately
		got, err := c.GetString(key)
		if err != nil {
			t.Errorf("GetString failed: %v", err)
		}
		if got != value {
			t.Errorf("GetString returned %v, want %v", got, value)
		}
	})

	// Test expiration
	t.Run("Expiration", func(t *testing.T) {
		key := "expiring_key"
		value := "expiring_value"
		expiration := 1 // 1 second

		// Set the value
		err := c.SetString(key, value, expiration)
		if err != nil {
			t.Errorf("SetString failed: %v", err)
		}

		// Wait for the key to expire
		time.Sleep(2 * time.Second)

		// Try to get the expired value
		_, err = c.GetString(key)
		if !errors.Is(err, ErrKeyExpired) {
			t.Errorf("Expected ErrKeyExpired, got %v", err)
		}
	})

	// Test non-existent key
	t.Run("Non-existent Key", func(t *testing.T) {
		key := "non_existent_key"

		_, err := c.GetString(key)
		if !errors.Is(err, ErrKeyNotFound) {
			t.Errorf("Expected ErrKeyNotFound, got %v", err)
		}
	})

	// Test overwriting an existing key
	t.Run("Overwrite Existing Key", func(t *testing.T) {
		key := "overwrite_key"
		value1 := "original_value"
		value2 := "new_value"
		expiration := 5 // 5 seconds

		// Set the original value
		err := c.SetString(key, value1, expiration)
		if err != nil {
			t.Errorf("SetString failed: %v", err)
		}

		// Overwrite with a new value
		err = c.SetString(key, value2, expiration)
		if err != nil {
			t.Errorf("SetString (overwrite) failed: %v", err)
		}

		// Get the value and check if it's the new one
		got, err := c.GetString(key)
		if err != nil {
			t.Errorf("GetString failed: %v", err)
		}
		if got != value2 {
			t.Errorf("GetString returned %v, want %v", got, value2)
		}
	})

	// Test setting a value with zero expiration
	t.Run("Zero Expiration", func(t *testing.T) {
		key := "zero_expiration_key"
		value := "zero_expiration_value"
		expiration := 0 // 0 seconds

		// Set the value with zero expiration
		err := c.SetString(key, value, expiration)
		if err != nil {
			t.Errorf("SetString failed: %v", err)
		}

		// Try to get the value immediately
		got, err := c.GetString(key)
		if err != nil {
			t.Errorf("GetString failed: %v", err)
		}
		if got != value {
			t.Errorf("GetString returned %v, want %v", got, value)
		}

		// Wait for a short time and try again
		time.Sleep(100 * time.Millisecond)
		got, err = c.GetString(key)
		if err != nil {
			t.Errorf("GetString failed after delay: %v", err)
		}
		if got != value {
			t.Errorf("GetString returned %v after delay, want %v", got, value)
		}
	})
}
