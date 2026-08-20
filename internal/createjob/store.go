package createjob

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Store struct {
	path string
}

func NewStore(path string) *Store {
	return &Store{path: path}
}

type storeFile struct {
	Version   int         `json:"version"`
	UpdatedAt time.Time   `json:"updated_at"`
	Jobs      []*Job      `json:"jobs"`
	Quota     *QuotaState `json:"quota,omitempty"`
}

type StoreState struct {
	Jobs  []*Job
	Quota QuotaState
}

func (s *Store) Load() ([]*Job, error) {
	state, err := s.LoadState()
	if err != nil {
		return nil, err
	}
	return state.Jobs, nil
}

func (s *Store) LoadState() (*StoreState, error) {
	raw, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &StoreState{Jobs: []*Job{}}, nil
		}
		return nil, err
	}

	var file storeFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, err
	}
	if file.Jobs == nil {
		file.Jobs = []*Job{}
	}
	state := &StoreState{Jobs: file.Jobs}
	if file.Quota != nil {
		state.Quota = *file.Quota
	}
	return state, nil
}

func (s *Store) Save(jobs []*Job) error {
	return s.SaveState(StoreState{Jobs: jobs})
}

func (s *Store) SaveState(state StoreState) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}
	if state.Jobs == nil {
		state.Jobs = []*Job{}
	}
	file := storeFile{
		Version:   1,
		UpdatedAt: time.Now(),
		Jobs:      state.Jobs,
	}
	if len(state.Quota.Buckets) > 0 {
		quota := state.Quota
		file.Quota = &quota
	}
	raw, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
