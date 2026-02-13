package session

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"
)

type State struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	UserID       string    `json:"userId"`
	SavedAt      time.Time `json:"savedAt"`
}

type Store struct {
	path string
	mu   sync.RWMutex
}

func NewStore(path string) *Store {
	return &Store{path: path}
}

func (s *Store) Load() (*State, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	content, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	if len(content) == 0 {
		return nil, nil
	}

	var state State
	if err := json.Unmarshal(content, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func (s *Store) Save(state *State) error {
	if state == nil {
		return errors.New("session state is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	state.SavedAt = time.Now().UTC()
	encoded, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, encoded, 0o600)
}

func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.Remove(s.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
