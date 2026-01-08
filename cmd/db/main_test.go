package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kalpovskii/checklist/internal/app/models"
	"github.com/kalpovskii/checklist/internal/app/pb"
	"github.com/kalpovskii/checklist/internal/app/services"
	"google.golang.org/protobuf/types/known/emptypb"
)

type mockTaskRepository struct {
	createFn   func(task *models.Task) error
	listFn     func() ([]models.Task, error)
	deleteFn   func(id uuid.UUID) error
	markDoneFn func(id uuid.UUID) error
}

func (m *mockTaskRepository) Create(task *models.Task) error {
	if m.createFn != nil {
		return m.createFn(task)
	}
	return nil
}

func (m *mockTaskRepository) List() ([]models.Task, error) {
	if m.listFn != nil {
		return m.listFn()
	}
	return []models.Task{}, nil
}

func (m *mockTaskRepository) Delete(id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(id)
	}
	return nil
}

func (m *mockTaskRepository) MarkDone(id uuid.UUID) error {
	if m.markDoneFn != nil {
		return m.markDoneFn(id)
	}
	return nil
}

type mockTaskCache struct {
	getTaskFn     func(ctx context.Context, id string) (*models.Task, error)
	setTaskFn     func(ctx context.Context, task *models.Task, ttl time.Duration) error
	getTaskListFn func(ctx context.Context) ([]models.Task, error)
	setTaskListFn func(ctx context.Context, tasks []models.Task, ttl time.Duration) error
	deleteTaskFn  func(ctx context.Context, id string) error
	deleteListFn  func(ctx context.Context) error
}

func (m *mockTaskCache) GetTask(ctx context.Context, id string) (*models.Task, error) {
	if m.getTaskFn != nil {
		return m.getTaskFn(ctx, id)
	}
	return nil, nil
}

func (m *mockTaskCache) SetTask(ctx context.Context, task *models.Task, ttl time.Duration) error {
	if m.setTaskFn != nil {
		return m.setTaskFn(ctx, task, ttl)
	}
	return nil
}

func (m *mockTaskCache) GetTaskList(ctx context.Context) ([]models.Task, error) {
	if m.getTaskListFn != nil {
		return m.getTaskListFn(ctx)
	}
	return nil, nil
}

func (m *mockTaskCache) SetTaskList(ctx context.Context, tasks []models.Task, ttl time.Duration) error {
	if m.setTaskListFn != nil {
		return m.setTaskListFn(ctx, tasks, ttl)
	}
	return nil
}

func (m *mockTaskCache) DeleteTask(ctx context.Context, id string) error {
	if m.deleteTaskFn != nil {
		return m.deleteTaskFn(ctx, id)
	}
	return nil
}

func (m *mockTaskCache) DeleteTaskList(ctx context.Context) error {
	if m.deleteListFn != nil {
		return m.deleteListFn(ctx)
	}
	return nil
}
func TestTaskServer_Create(t *testing.T) {
	t.Run("успешное создание задачи", func(t *testing.T) {
		taskID := uuid.New()
		createdAt := time.Now()

		mockRepo := &mockTaskRepository{
			createFn: func(task *models.Task) error {
				task.ID = taskID
				task.CreatedAt = createdAt
				return nil
			},
		}

		mockCache := &mockTaskCache{}

		service := services.NewTaskService(mockRepo, mockCache)

		server := &TaskServer{
			service: service,
		}

		req := &pb.CreateTaskRequest{
			Title:   "Тестовая задача",
			Content: "Описание задачи",
		}

		resp, err := server.Create(context.Background(), req)

		if err != nil {
			t.Fatalf("неожиданная ошибка: %v", err)
		}

		if resp == nil || resp.Task == nil {
			t.Fatal("ответ не должен быть nil")
		}

		if resp.Task.Id != taskID.String() {
			t.Errorf("неожиданный ID: ожидалось %s, получено %s", taskID.String(), resp.Task.Id)
		}

		if resp.Task.Title != "Тестовая задача" {
			t.Errorf("неожиданный Title: ожидалось 'Тестовая задача', получено '%s'", resp.Task.Title)
		}

		if resp.Task.Content != "Описание задачи" {
			t.Errorf("неожиданный Content: ожидалось 'Описание задачи', получено '%s'", resp.Task.Content)
		}

		if resp.Task.Done != false {
			t.Errorf("неожиданный Done: ожидалось false, получено %v", resp.Task.Done)
		}

		if resp.Task.CreatedAt == nil {
			t.Error("CreatedAt не должен быть nil")
		}
	})

	t.Run("ошибка при создании задачи", func(t *testing.T) {
		expectedError := errors.New("ошибка базы данных")

		mockRepo := &mockTaskRepository{
			createFn: func(task *models.Task) error {
				return expectedError
			},
		}

		mockCache := &mockTaskCache{}

		service := services.NewTaskService(mockRepo, mockCache)
		server := &TaskServer{
			service: service,
		}

		req := &pb.CreateTaskRequest{
			Title:   "Задача",
			Content: "Описание",
		}

		resp, err := server.Create(context.Background(), req)

		if err == nil {
			t.Fatal("ожидалась ошибка, но её не было")
		}

		if resp != nil {
			t.Error("ответ должен быть nil при ошибке")
		}

		if err.Error() != expectedError.Error() {
			t.Errorf("неожиданная ошибка: ожидалось '%v', получено '%v'", expectedError, err)
		}
	})
}

func TestTaskServer_List(t *testing.T) {
	t.Run("успешное получение списка задач", func(t *testing.T) {
		taskID1 := uuid.New()
		taskID2 := uuid.New()
		createdAt := time.Now()

		expectedTasks := []models.Task{
			{
				ID:        taskID1,
				Title:     "Задача 1",
				Content:   "Описание 1",
				Done:      false,
				CreatedAt: createdAt,
			},
			{
				ID:        taskID2,
				Title:     "Задача 2",
				Content:   "Описание 2",
				Done:      true,
				CreatedAt: createdAt.Add(time.Hour),
			},
		}

		mockRepo := &mockTaskRepository{
			listFn: func() ([]models.Task, error) {
				return expectedTasks, nil
			},
		}

		mockCache := &mockTaskCache{
			getTaskListFn: func(ctx context.Context) ([]models.Task, error) {
				return nil, nil
			},
		}

		service := services.NewTaskService(mockRepo, mockCache)
		server := &TaskServer{
			service: service,
		}

		resp, err := server.List(context.Background(), &emptypb.Empty{})

		if err != nil {
			t.Fatalf("неожиданная ошибка: %v", err)
		}

		if resp == nil {
			t.Fatal("ответ не должен быть nil")
		}

		if len(resp.Tasks) != 2 {
			t.Fatalf("неожиданное количество задач: ожидалось 2, получено %d", len(resp.Tasks))
		}

		if resp.Tasks[0].Id != taskID1.String() {
			t.Errorf("неожиданный ID первой задачи: ожидалось %s, получено %s", taskID1.String(), resp.Tasks[0].Id)
		}
		if resp.Tasks[0].Title != "Задача 1" {
			t.Errorf("неожиданный Title первой задачи: ожидалось 'Задача 1', получено '%s'", resp.Tasks[0].Title)
		}
		if resp.Tasks[0].Done != false {
			t.Errorf("неожиданный Done первой задачи: ожидалось false, получено %v", resp.Tasks[0].Done)
		}

		if resp.Tasks[1].Id != taskID2.String() {
			t.Errorf("неожиданный ID второй задачи: ожидалось %s, получено %s", taskID2.String(), resp.Tasks[1].Id)
		}
		if resp.Tasks[1].Title != "Задача 2" {
			t.Errorf("неожиданный Title второй задачи: ожидалось 'Задача 2', получено '%s'", resp.Tasks[1].Title)
		}
		if resp.Tasks[1].Done != true {
			t.Errorf("неожиданный Done второй задачи: ожидалось true, получено %v", resp.Tasks[1].Done)
		}
	})

	t.Run("пустой список задач", func(t *testing.T) {
		mockRepo := &mockTaskRepository{
			listFn: func() ([]models.Task, error) {
				return []models.Task{}, nil
			},
		}

		mockCache := &mockTaskCache{
			getTaskListFn: func(ctx context.Context) ([]models.Task, error) {
				return nil, nil
			},
		}

		service := services.NewTaskService(mockRepo, mockCache)
		server := &TaskServer{
			service: service,
		}

		resp, err := server.List(context.Background(), &emptypb.Empty{})

		if err != nil {
			t.Fatalf("неожиданная ошибка: %v", err)
		}

		if resp == nil {
			t.Fatal("ответ не должен быть nil")
		}

		if len(resp.Tasks) != 0 {
			t.Errorf("неожиданное количество задач: ожидалось 0, получено %d", len(resp.Tasks))
		}
	})

	t.Run("ошибка при получении списка", func(t *testing.T) {
		expectedError := errors.New("ошибка базы данных")

		mockRepo := &mockTaskRepository{
			listFn: func() ([]models.Task, error) {
				return nil, expectedError
			},
		}

		mockCache := &mockTaskCache{
			getTaskListFn: func(ctx context.Context) ([]models.Task, error) {
				return nil, nil
			},
		}

		service := services.NewTaskService(mockRepo, mockCache)
		server := &TaskServer{
			service: service,
		}

		resp, err := server.List(context.Background(), &emptypb.Empty{})

		if err == nil {
			t.Fatal("ожидалась ошибка, но её не было")
		}

		if resp != nil {
			t.Error("ответ должен быть nil при ошибке")
		}
	})
}

func TestTaskServer_Delete(t *testing.T) {
	t.Run("успешное удаление задачи", func(t *testing.T) {
		taskID := uuid.New()

		mockRepo := &mockTaskRepository{
			deleteFn: func(id uuid.UUID) error {
				if id != taskID {
					t.Errorf("неожиданный ID: ожидалось %s, получено %s", taskID, id)
				}
				return nil
			},
		}

		mockCache := &mockTaskCache{}

		service := services.NewTaskService(mockRepo, mockCache)
		server := &TaskServer{
			service: service,
		}

		req := &pb.TaskIDRequest{
			Id: taskID.String(),
		}

		resp, err := server.Delete(context.Background(), req)

		if err != nil {
			t.Fatalf("неожиданная ошибка: %v", err)
		}

		if resp == nil {
			t.Fatal("ответ не должен быть nil")
		}

		if resp.Status != "deleted" {
			t.Errorf("неожиданный статус: ожидалось 'deleted', получено '%s'", resp.Status)
		}
	})

	t.Run("ошибка парсинга UUID", func(t *testing.T) {
		mockRepo := &mockTaskRepository{}
		mockCache := &mockTaskCache{}

		service := services.NewTaskService(mockRepo, mockCache)
		server := &TaskServer{
			service: service,
		}

		req := &pb.TaskIDRequest{
			Id: "невалидный-uuid",
		}

		resp, err := server.Delete(context.Background(), req)

		if err == nil {
			t.Fatal("ожидалась ошибка парсинга UUID, но её не было")
		}

		if resp != nil {
			t.Error("ответ должен быть nil при ошибке")
		}

		_, parseErr := uuid.Parse("невалидный-uuid")
		if parseErr == nil {
			t.Error("uuid.Parse должен вернуть ошибку для невалидного UUID")
		}
	})

	t.Run("ошибка при удалении задачи", func(t *testing.T) {
		taskID := uuid.New()
		expectedError := errors.New("задача не найдена")

		mockRepo := &mockTaskRepository{
			deleteFn: func(id uuid.UUID) error {
				return expectedError
			},
		}

		mockCache := &mockTaskCache{}

		service := services.NewTaskService(mockRepo, mockCache)
		server := &TaskServer{
			service: service,
		}

		req := &pb.TaskIDRequest{
			Id: taskID.String(),
		}

		resp, err := server.Delete(context.Background(), req)

		if err == nil {
			t.Fatal("ожидалась ошибка, но её не было")
		}

		if resp != nil {
			t.Error("ответ должен быть nil при ошибке")
		}
	})
}

func TestTaskServer_MarkDone(t *testing.T) {
	t.Run("успешная отметка задачи как выполненной", func(t *testing.T) {
		taskID := uuid.New()

		mockRepo := &mockTaskRepository{
			markDoneFn: func(id uuid.UUID) error {
				if id != taskID {
					t.Errorf("неожиданный ID: ожидалось %s, получено %s", taskID, id)
				}
				return nil
			},
		}

		mockCache := &mockTaskCache{}

		service := services.NewTaskService(mockRepo, mockCache)
		server := &TaskServer{
			service: service,
		}

		req := &pb.TaskIDRequest{
			Id: taskID.String(),
		}

		resp, err := server.MarkDone(context.Background(), req)

		if err != nil {
			t.Fatalf("неожиданная ошибка: %v", err)
		}

		if resp == nil {
			t.Fatal("ответ не должен быть nil")
		}

		if resp.Status != "done" {
			t.Errorf("неожиданный статус: ожидалось 'done', получено '%s'", resp.Status)
		}
	})

	t.Run("ошибка парсинга UUID", func(t *testing.T) {
		mockRepo := &mockTaskRepository{}
		mockCache := &mockTaskCache{}

		service := services.NewTaskService(mockRepo, mockCache)
		server := &TaskServer{
			service: service,
		}

		req := &pb.TaskIDRequest{
			Id: "невалидный-uuid",
		}

		resp, err := server.MarkDone(context.Background(), req)

		if err == nil {
			t.Fatal("ожидалась ошибка парсинга UUID, но её не было")
		}

		if resp != nil {
			t.Error("ответ должен быть nil при ошибке")
		}
	})

	t.Run("ошибка при отметке задачи", func(t *testing.T) {
		taskID := uuid.New()
		expectedError := errors.New("задача не найдена")

		mockRepo := &mockTaskRepository{
			markDoneFn: func(id uuid.UUID) error {
				return expectedError
			},
		}

		mockCache := &mockTaskCache{}

		service := services.NewTaskService(mockRepo, mockCache)
		server := &TaskServer{
			service: service,
		}

		req := &pb.TaskIDRequest{
			Id: taskID.String(),
		}

		resp, err := server.MarkDone(context.Background(), req)

		if err == nil {
			t.Fatal("ожидалась ошибка, но её не было")
		}

		if resp != nil {
			t.Error("ответ должен быть nil при ошибке")
		}
	})
}
