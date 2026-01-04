package services

import (
	"github.com/google/uuid"
	"github.com/kalpovskii/checklist/internal/app/models"
	"github.com/kalpovskii/checklist/internal/app/repositories"
)

type TaskService struct {
	repo repositories.TaskRepository
}

func NewTaskService(repo repositories.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) Create(title, content string) (*models.Task, error) {
	task := &models.Task{Title: title, Content: content}
	err := s.repo.Create(task)
	return task, err
}

func (s *TaskService) List() ([]models.Task, error) {
	return s.repo.List()
}

func (s *TaskService) Delete(id uuid.UUID) error {
	return s.repo.Delete(id)
}

func (s *TaskService) MarkDone(id uuid.UUID) error {
	return s.repo.MarkDone(id)
}