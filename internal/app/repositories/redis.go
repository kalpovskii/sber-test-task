package repositories

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kalpovskii/checklist/internal/app/models"
	"github.com/redis/go-redis/v9"
)

type TaskCache interface {
	GetTask(ctx context.Context, id string) (*models.Task, error)
	SetTask(ctx context.Context, task *models.Task, ttl time.Duration) error

	GetTaskList(ctx context.Context) ([]models.Task, error)
	SetTaskList(ctx context.Context, tasks []models.Task, ttl time.Duration) error

	DeleteTask(ctx context.Context, id string) error
	DeleteTaskList(ctx context.Context) error
}

type RedisTaskRepository struct {
	rdb *redis.Client
}

func NewRedisTaskRepository(rdb *redis.Client) *RedisTaskRepository {
	return &RedisTaskRepository{rdb: rdb}
}

func taskKey(id string) string {
	return "task:" + id
}

const taskListKey = "tasks:list"

func (r *RedisTaskRepository) GetTask(
	ctx context.Context,
	id string,
) (*models.Task, error) {

	val, err := r.rdb.Get(ctx, taskKey(id)).Result()
	if err == redis.Nil {
		return nil, nil // cache miss
	}
	if err != nil {
		return nil, err
	}

	var task models.Task
	if err := json.Unmarshal([]byte(val), &task); err != nil {
		return nil, err
	}

	return &task, nil
}

func (r *RedisTaskRepository) SetTask(
	ctx context.Context,
	task *models.Task,
	ttl time.Duration,
) error {

	data, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return r.rdb.Set(ctx, taskKey(task.ID.String()), data, ttl).Err()
}

func (r *RedisTaskRepository) DeleteTask(ctx context.Context, id string) error {
	return r.rdb.Del(ctx, taskKey(id)).Err()
}

func (r *RedisTaskRepository) DeleteTaskList(ctx context.Context) error {
	return r.rdb.Del(ctx, taskListKey).Err()
}

func (r *RedisTaskRepository) GetTaskList(ctx context.Context) ([]models.Task, error) {
	val, err := r.rdb.Get(ctx, taskListKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var tasks []models.Task
	if err := json.Unmarshal([]byte(val), &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *RedisTaskRepository) SetTaskList(
	ctx context.Context,
	tasks []models.Task,
	ttl time.Duration,
) error {

	data, err := json.Marshal(tasks)
	if err != nil {
		return err
	}

	return r.rdb.Set(ctx, taskListKey, data, ttl).Err()
}

