package notes

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	storage Storage
}

func NewService(storage Storage) *Service {
	return &Service{storage: storage}
}

func (s *Service) List() ([]Note, error) {
	return s.storage.GetAll()
}

func (s *Service) Get(id string) (*Note, error) {
	if id == "" {
		return nil, fmt.Errorf("note not found")
	}
	return s.storage.GetByID(id)
}

func (s *Service) Create(title, content, template string) (*Note, error) {
	if title == "" && content == "" {
		return nil, fmt.Errorf("title and content cannot both be empty")
	}

	note := &Note{
		ID:        uuid.New().String(),
		Title:     title,
		Content:   content,
		Template:  template,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	if err := s.storage.Save(note); err != nil {
		return nil, err
	}

	return note, nil
}

func (s *Service) Update(id, title, content, template string) (*Note, error) {
	existing, err := s.storage.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("note not found")
	}

	if title == "" && content == "" {
		return nil, fmt.Errorf("title and content cannot both be empty")
	}

	existing.Title = title
	existing.Content = content
	existing.Template = template

	if err := s.storage.Update(existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *Service) Delete(id string) error {
	return s.storage.Delete(id)
}