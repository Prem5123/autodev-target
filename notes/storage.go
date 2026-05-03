package notes

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Note struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Template  string `json:"template"`
	CreatedAt string `json:"created_at"`
}

type Storage interface {
	GetAll() ([]Note, error)
	GetByID(id string) (*Note, error)
	Save(note *Note) error
	Update(note *Note) error
	Delete(id string) error
}

type FileStorage struct {
	filename string
	mu       sync.RWMutex
	notes    map[string]Note
}

func NewFileStorage(filename string) (*FileStorage, error) {
	storage := &FileStorage{
		filename: filename,
		notes:    make(map[string]Note),
	}

	// Load existing notes from file
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return storage, nil
		}
		return nil, fmt.Errorf("failed to open storage file: %w", err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&storage.notes); err != nil {
		return nil, fmt.Errorf("failed to decode storage: %w", err)
	}

	return storage, nil
}

func (s *FileStorage) persist() error {
	file, err := os.Create(s.filename)
	if err != nil {
		return fmt.Errorf("failed to create storage file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(s.notes)
}

func (s *FileStorage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.persist()
}

func (s *FileStorage) GetAll() ([]Note, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	notes := make([]Note, 0, len(s.notes))
	for _, note := range s.notes {
		notes = append(notes, note)
	}
	return notes, nil
}

func (s *FileStorage) GetByID(id string) (*Note, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	note, exists := s.notes[id]
	if !exists {
		return nil, nil
	}
	return &note, nil
}

func (s *FileStorage) Save(note *Note) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.notes[note.ID] = *note
	return s.persist()
}

func (s *FileStorage) Update(note *Note) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.notes[note.ID]; !exists {
		return fmt.Errorf("note not found")
	}

	s.notes[note.ID] = *note
	return s.persist()
}

func (s *FileStorage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.notes[id]; !exists {
		return fmt.Errorf("note not found")
	}

	delete(s.notes, id)
	return s.persist()
}