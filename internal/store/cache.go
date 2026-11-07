package store

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"jellyfin-lolesports-streams/internal/types"
	"sync"
	"time"
)

type Cache struct {
	guideCache guideCache
	imageCache map[string][]byte
}

type guideCache struct {
	mu        sync.RWMutex
	guide     []byte
	hash      [32]byte
	createdAt time.Time
}

func (c *Cache) GetImageCache(key string) ([]byte, bool) {
	cacheEntry := c.imageCache[key]

	if cacheEntry == nil {
		return nil, false
	}
	return cacheEntry, true
}

func (c *Cache) SetImageCache(image []byte, key string) {
	c.imageCache[key] = image
}

func (c *Cache) GetGuideCache(channels []types.Channel, ttl time.Duration) ([]byte, bool) {
	h, err := hashChannels(channels)
	if err != nil {
		return nil, false
	}

	c.guideCache.mu.RLock()
	defer c.guideCache.mu.RUnlock()

	if c.guideCache.guide == nil {
		return nil, false
	}

	if time.Since(c.guideCache.createdAt) >= ttl {
		return nil, false
	}

	if h != c.guideCache.hash {
		return nil, false
	}

	return c.guideCache.guide, true
}

func (c *Cache) SetGuideCache(guide []byte, channels []types.Channel) {
	h, err := hashChannels(channels)
	if err != nil {
		return
	}

	c.guideCache.mu.Lock()
	defer c.guideCache.mu.Unlock()

	c.guideCache.hash = h
	c.guideCache.guide = guide
	c.guideCache.createdAt = time.Now()
}

func hashChannels(channels []types.Channel) ([32]byte, error) {
	var bytesBuffer bytes.Buffer
	err := gob.NewEncoder(&bytesBuffer).Encode(channels)
	if err != nil {
		return [32]byte{}, err
	}

	h := sha256.Sum256(bytesBuffer.Bytes())

	return h, nil
}
