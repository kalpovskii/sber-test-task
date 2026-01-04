package repositories

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/kalpovskii/checklist/internal/app/models"
	_ "github.com/lib/pq"
)

type TaskRepository interface {
	Create(task *models.Task) error
	List() ([]models.Task, error)
	Delete(id uuid.UUID) error
	MarkDone(id uuid.UUID) error
}

type PostgresTaskRepo struct {
	db *sql.DB
}

func NewPostgresTaskRepo(dsn string) (*PostgresTaskRepo, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id UUID PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT,
			done BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return nil, err
	}

	return &PostgresTaskRepo{db: db}, nil
}

func (r *PostgresTaskRepo) Create(task *models.Task) error {
	task.ID = uuid.New()
	task.CreatedAt = time.Now()
	_, err := r.db.Exec("INSERT INTO tasks (id, title, content, done, created_at) VALUES ($1, $2, $3, $4, $5)",
		task.ID, task.Title, task.Content, task.Done, task.CreatedAt)
	return err
}

func (r *PostgresTaskRepo) List() ([]models.Task, error) {
	rows, err := r.db.Query("SELECT id, title, content, done, created_at FROM tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		err := rows.Scan(&t.ID, &t.Title, &t.Content, &t.Done, &t.CreatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *PostgresTaskRepo) Delete(id uuid.UUID) error {
	_, err := r.db.Exec("DELETE FROM tasks WHERE id = $1", id)
	return err
}

func (r *PostgresTaskRepo) MarkDone(id uuid.UUID) error {
	_, err := r.db.Exec("UPDATE tasks SET done = TRUE WHERE id = $1", id)
	return err
}