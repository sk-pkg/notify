// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package msgid

import (
	"github.com/sk-pkg/notify/util"
	"log"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	initIndex = 10000000 // Initial sequence number
	indexBase = 36       // Base for sequence number conversion
)

var (
	hostnameOnce sync.Once // Ensures hostname is retrieved only once
	hostname     string    // Cached hostname
)

// ID represents a structure for generating unique identifiers.
// It uses a combination of hostname, timestamp, and a sequence number
// to ensure uniqueness across distributed systems.
type ID struct {
	index  uint64     // Sequence number, accessed atomically
	prefix string     // Prefix containing timestamp and hostname
	mu     sync.Mutex // Mutex to ensure thread-safety when updating the prefix
}

// NewMessageID creates and returns a New ID instance.
// It initializes the ID with the current hostname and timestamp.
//
// Returns:
//   - *ID: A pointer to the newly created ID instance.
//
// Example:
//
//	messageID := NewMessageID()
//	uniqueIdentifier := messageID.New()
func NewMessageID() *ID {
	t := &ID{
		index: initIndex,
	}
	t.updatePrefix()
	return t
}

// updatePrefix combines the current timestamp and cached hostname to form the prefix.
// This method should be called with external synchronization.
func (m *ID) updatePrefix() {
	var err error

	m.mu.Lock()
	defer m.mu.Unlock()

	// Retrieve hostname only once
	hostnameOnce.Do(func() {
		hostname, err = os.Hostname()
		if err != nil {
			log.Printf("Failed to get hostname: %v", err)
			// Use a default value if hostname retrieval fails
			hostname = "unknown"
		}
	})

	// Construct the prefix using hostname and current timestamp
	m.prefix = util.SpliceStr(hostname, "-", strconv.FormatInt(time.Now().UnixNano(), indexBase), "-")
	m.index = initIndex
}

// New generates and returns a New unique identifier.
//
// Returns:
//   - string: A unique identifier combining the prefix and a sequence number.
//
// Example:
//
//	messageID := NewMessageID()
//	id1 := messageID.New() // e.g., "hostname-timestamp-sequence1"
//	id2 := messageID.New() // e.g., "hostname-timestamp-sequence2"
func (m *ID) New() string {
	// Atomically increment the sequence number
	newIndex := atomic.AddUint64(&m.index, 1)

	// If the sequence number overflows, update the prefix and reset the sequence
	if newIndex == 0 {
		m.mu.Lock()
		defer m.mu.Unlock()
		if atomic.LoadUint64(&m.index) == 0 {
			m.updatePrefix()
		}
	}

	// Convert the sequence number to a base-36 string
	id := strconv.FormatUint(newIndex, indexBase)

	// Combine the prefix and the sequence number
	return util.SpliceStr(m.prefix, id)
}
