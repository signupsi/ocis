package roles

import (
	"sync"
	"time"

	settings "github.com/owncloud/ocis-settings/pkg/proto/v0"
)

// entry extends a bundle and adds a TTL
type entry struct {
	*settings.Bundle
	inserted time.Time
}

// Cache is a barebones cache implementation.
type Cache struct {
	entries map[string]entry
	size    int
	ttl     time.Duration
	m       sync.Mutex
}

// NewCache returns a new instance of Cache.
func NewCache(o ...Option) Cache {
	opts := newOptions(o...)

	return Cache{
		size:    opts.size,
		ttl:	 opts.ttl,
		entries: map[string]entry{},
	}
}

// Get gets a role-bundle by a given `roleID`.
func (c *Cache) Get(roleID string) *settings.Bundle {
	c.m.Lock()
	defer c.m.Unlock()

	if _, ok := c.entries[roleID]; ok {
		return c.entries[roleID].Bundle
	}
	return nil
}

// FindPermissionById searches for a setting with the permissionID, but limited to the given roleIDs
func (c *Cache) FindPermissionById(roleIDs []string, permissionID string) *settings.Setting {
	for _, roleID := range roleIDs {
		role := c.Get(roleID)
		if role != nil {
			for _, setting := range role.Settings {
				if setting.Id == permissionID {
					return setting
				}
			}
		}
	}
	return nil
}

// Set sets a roleID / role-bundle.
func (c *Cache) Set(roleID string, value *settings.Bundle) {
	c.m.Lock()
	defer c.m.Unlock()

	if !c.fits() {
		c.evict()
	}

	c.entries[roleID] = entry{
		value,
		time.Now(),
	}
}

// Evict frees memory from the cache by removing invalid keys. It is a noop.
func (c *Cache) evict() {
	for i := range c.entries {
		if c.entries[i].inserted.Add(c.ttl).Before(time.Now()) {
			delete(c.entries, i)
		}
	}
}

// Length returns the amount of entries per service key.
func (c *Cache) Length() int {
	return len(c.entries)
}

func (c *Cache) fits() bool {
	return c.size >= len(c.entries)
}
