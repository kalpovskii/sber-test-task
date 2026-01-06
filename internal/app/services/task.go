package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kalpovskii/checklist/internal/app/models"
	"github.com/kalpovskii/checklist/internal/app/repositories"
)

const (
	taskTTL     = 60 * time.Second
	taskListTTL = 15 * time.Second
)

type TaskService struct {
	repo  repositories.TaskRepository
	cache repositories.TaskCache
}

func NewTaskService(repo repositories.TaskRepository, cache repositories.TaskCache) *TaskService {
	return &TaskService{
		repo:  repo,
		cache: cache,
	}
}

func (s *TaskService) Create(title, content string) (*models.Task, error) {
	task := &models.Task{
		Title:   title,
		Content: content,
	}

	if err := s.repo.Create(task); err != nil {
		return nil, err
	}

	ctx := context.Background()

	_ = s.cache.SetTask(ctx, task, taskTTL)
	_ = s.cache.DeleteTaskList(ctx)

	return task, nil
}

func (s *TaskService) List() ([]models.Task, error) {
	ctx := context.Background()

	if tasks, err := s.cache.GetTaskList(ctx); err == nil && tasks != nil {
		return tasks, nil
	}

	tasks, err := s.repo.List()
	if err != nil {
		return nil, err
	}

	_ = s.cache.SetTaskList(ctx, tasks, taskListTTL)

	return tasks, nil
}

func (s *TaskService) Delete(id uuid.UUID) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}

	ctx := context.Background()

	_ = s.cache.DeleteTask(ctx, id.String())
	_ = s.cache.DeleteTaskList(ctx)

	return nil
}

func (s *TaskService) MarkDone(id uuid.UUID) error {
	if err := s.repo.MarkDone(id); err != nil {
		return err
	}

	ctx := context.Background()

	_ = s.cache.DeleteTask(ctx, id.String())
	_ = s.cache.DeleteTaskList(ctx)

	return nil
}