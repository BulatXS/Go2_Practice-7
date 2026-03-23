package service

import (
	"context"
	"fmt"
	"sync/atomic"

	"singularity.com/pr7/services/tasks/internal/repository"
)

type TaskService struct {
	repo    repository.TaskRepository
	counter uint64
}

func NewTaskService(repo repository.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) Create(ctx context.Context, title, description, dueDate string) (Task, error) {
	id := fmt.Sprintf("t_%03d", atomic.AddUint64(&s.counter, 1))

	task := repository.Task{
		ID:          id,
		Title:       title,
		Description: description,
		DueDate:     dueDate,
		Done:        false,
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return Task{}, err
	}

	return Task{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		DueDate:     task.DueDate,
		Done:        task.Done,
	}, nil
}

func (s *TaskService) GetAll(ctx context.Context) ([]Task, error) {
	items, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]Task, 0, len(items))
	for _, t := range items {
		result = append(result, Task{
			ID:          t.ID,
			Title:       t.Title,
			Description: t.Description,
			DueDate:     t.DueDate,
			Done:        t.Done,
		})
	}

	return result, nil
}

func (s *TaskService) GetByID(ctx context.Context, id string) (Task, bool, error) {
	t, ok, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Task{}, false, err
	}
	if !ok {
		return Task{}, false, nil
	}

	return Task{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		DueDate:     t.DueDate,
		Done:        t.Done,
	}, true, nil
}

func (s *TaskService) Update(
	ctx context.Context,
	id string,
	title *string,
	description *string,
	dueDate *string,
	done *bool,
) (Task, bool, error) {
	t, ok, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Task{}, false, err
	}
	if !ok {
		return Task{}, false, nil
	}

	if title != nil {
		t.Title = *title
	}
	if description != nil {
		t.Description = *description
	}
	if dueDate != nil {
		t.DueDate = *dueDate
	}
	if done != nil {
		t.Done = *done
	}

	if err := s.repo.Update(ctx, t); err != nil {
		return Task{}, false, err
	}

	return Task{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		DueDate:     t.DueDate,
		Done:        t.Done,
	}, true, nil
}

func (s *TaskService) Delete(ctx context.Context, id string) (bool, error) {
	return s.repo.Delete(ctx, id)
}

func (s *TaskService) SearchByTitleSafe(ctx context.Context, title string) ([]Task, error) {
	items, err := s.repo.SearchByTitleSafe(ctx, title)
	if err != nil {
		return nil, err
	}

	result := make([]Task, 0, len(items))
	for _, t := range items {
		result = append(result, Task{
			ID:          t.ID,
			Title:       t.Title,
			Description: t.Description,
			DueDate:     t.DueDate,
			Done:        t.Done,
		})
	}

	return result, nil
}
