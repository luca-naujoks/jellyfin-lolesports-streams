package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"jellyfin-lolesports-streams/internal/types"
	"sync"
	"time"
)

type GuideCache struct {
	mu        sync.RWMutex
	guide     []byte
	hash      [32]byte
	createdAt time.Time
}

func (g *GuideCache) Get(channels []types.Channel, ttl time.Duration) ([]byte, bool) {
	h, err := hashChannels(channels)
	if err != nil {
		return nil, false
	}

	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.guide == nil {
		return nil, false
	}

	if time.Since(g.createdAt) >= ttl {
		return nil, false
	}

	if h != g.hash {
		return nil, false
	}

	return g.guide, true
}

func (g *GuideCache) Set(guide []byte, channels []types.Channel) {
	h, err := hashChannels(channels)
	if err != nil {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	g.hash = h
	g.guide = guide
	g.createdAt = time.Now()
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
