package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/iancenry/jarvis/internal/model/todo"
	"github.com/jackc/pgx/v5"
)

func (r *TodoRepository) GetTodosDueInHours(ctx context.Context, hours, limit int) ([]todo.Todo, error) {
	stmt := `
		SELECT t.* FROM todos t
		WHERE t.due_date IS NOT NULL
			AND t.due_date >= NOW()
			AND t.due_date <= NOW() + (@hours * INTERVAL '1 hour')
			AND t.status NOT IN ('completed', 'archived')
		ORDER BY t.due_date ASC
		LIMIT @limit
	`

	rows, err := r.db.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"hours": hours,
		"limit": limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get todos due in hours query for hours=%d: %w", hours, err)
	}

	todos, err := pgx.CollectRows(rows, pgx.RowToStructByName[todo.Todo])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []todo.Todo{}, nil
		}
		return nil, fmt.Errorf("failed to collect todos due in hours for hours=%d: %w", hours, err)
	}

	return todos, nil
}

func (r *TodoRepository) GetOverdueTodos(ctx context.Context, limit int) ([]todo.Todo, error) {
	stmt := `
		SELECT t.* FROM todos t
		WHERE t.due_date IS NOT NULL
			AND t.due_date < NOW()
			AND t.status NOT IN ('completed', 'archived')
		ORDER BY t.due_date ASC
		LIMIT @limit
	`

	rows, err := r.db.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"limit": limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get overdue todos query: %w", err)
	}

	todos, err := pgx.CollectRows(rows, pgx.RowToStructByName[todo.Todo])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []todo.Todo{}, nil
		}
		return nil, fmt.Errorf("failed to collect overdue todos: %w", err)
	}

	return todos, nil
}

func (r *TodoRepository) GetCompletedTodosOlderThan(ctx context.Context, cutoffDate time.Time, limit int) ([]todo.Todo, error) {
	stmt := `
		SELECT *
		FROM todos
		WHERE status = 'completed'
			AND completed_at IS NOT NULL
			AND completed_at < @cutoff_date
		ORDER BY completed_at ASC
		LIMIT @limit
	`

	rows, err := r.db.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"cutoff_date": cutoffDate,
		"limit":       limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get completed todos older than %s query: %w", cutoffDate.Format("2006-01-02"), err)
	}

	todos, err := pgx.CollectRows(rows, pgx.RowToStructByName[todo.Todo])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []todo.Todo{}, nil
		}
		return nil, fmt.Errorf("failed to collect rows from table:todos: %w", err)
	}

	return todos, nil
}

func (r *TodoRepository) ArchiveTodos(ctx context.Context, todoIDs []uuid.UUID) error {
	stmt := `
		UPDATE todos
		SET status = 'archived'
		WHERE id = ANY(@ids) AND status != 'archived'
	`

	result, err := r.db.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"ids": todoIDs,
	})
	if err != nil {
		return fmt.Errorf("failed to execute archive todos query: %w", err)
	}

	if result.RowsAffected() != int64(len(todoIDs)) {
		return fmt.Errorf("not all todos were archived, expected to archive %d but archived %d", len(todoIDs), result.RowsAffected())
	}

	return nil
}
