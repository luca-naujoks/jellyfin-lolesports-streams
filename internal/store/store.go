package store

import (
	"errors"
	"fmt"
	"jellyfin-lolesports-streams/internal/types"
	"maps"
	"slices"
	"sync"
)

type Store struct {
	mutex    sync.RWMutex
	channels []types.Channel
}

var ErrChannelNotFound = errors.New("channel not found")

func NewStore() *Store {
	return &Store{
		channels: make([]types.Channel, 0),
	}
}

func cloneChannel(ch types.Channel) types.Channel {
	ch.Leagues = slices.Clone(ch.Leagues)

	return ch
}

func (s *Store) GetAll() []types.Channel {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	channels := make([]types.Channel, len(s.channels))

	for i, ch := range s.channels {
		channels[i] = cloneChannel(ch)
	}

	return channels
}

func (s *Store) GetByID(id string) (types.Channel, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for i := range s.channels {
		if s.channels[i].ID == id {
			return cloneChannel(s.channels[i]), nil
		}
	}

	return types.Channel{}, fmt.Errorf("%w: %s", ErrChannelNotFound, id)
}

func (s *Store) Add(channel types.Channel) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, ch := range s.channels {
		if ch.ID == channel.ID {
			return fmt.Errorf("a channel with ID %s already exists", channel.ID)
		}
	}

	s.channels = append(s.channels, cloneChannel(channel))

	return nil
}

func (s *Store) Update(channel types.Channel) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i := range s.channels {
		if s.channels[i].ID == channel.ID {
			s.channels[i] = cloneChannel(channel)
			return nil
		}
	}

	return fmt.Errorf("%w: %s", ErrChannelNotFound, channel.ID)
}
func (s *Store) Delete(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i := range s.channels {
		if s.channels[i].ID == id {
			s.channels = slices.Delete(s.channels, i, i+1)
			return nil
		}
	}

	return fmt.Errorf("%w: %s", ErrChannelNotFound, id)
}

func (s *Store) GetLeaguesUsedByChannels() []types.League {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	uniqueLeagues := make(map[string]types.League)

	for _, ch := range s.channels {
		for _, league := range ch.Leagues {
			uniqueLeagues[league.Slug] = league
		}
	}

	return slices.Collect(maps.Values(uniqueLeagues))
}
