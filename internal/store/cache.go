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
	imageCache imageCache
}

type imageCache struct {
	mu     sync.RWMutex
	images map[string][]byte
}

type guideCache struct {
	mu        sync.RWMutex
	guide     []byte
	hash      [32]byte
	createdAt time.Time
}

func (c *Cache) GetImageCache(key string) ([]byte, bool) {
	c.imageCache.mu.RLock()
	defer c.imageCache.mu.RUnlock()

	cacheEntry := c.imageCache.images[key]

	if cacheEntry == nil {
		return nil, false
	}

	return cacheEntry, true
}

func (c *Cache) SetImageCache(image []byte, key string) {
	c.imageCache.mu.Lock()
	defer c.imageCache.mu.Unlock()

	if c.imageCache.images == nil {
		c.imageCache.images = make(map[string][]byte)
	}

	c.imageCache.images[key] = image
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
	c.guideCache.mu.Lock()
	defer c.guideCache.mu.Unlock()

	if c.guideCache.guide == nil {
		c.guideCache = guideCache{
			guide:     nil,
			hash:      [32]byte{},
			createdAt: time.Time{},
		}
	}

	h, err := hashChannels(channels)
	if err != nil {
		return
	}

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
