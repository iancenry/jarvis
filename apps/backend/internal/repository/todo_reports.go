package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iancenry/jarvis/internal/model/todo"
	"github.com/jackc/pgx/v5"
)

func (r *TodoRepository) GetWeeklyStatsForUsers(ctx context.Context, startDate, endDate time.Time) ([]todo.UserWeeklyStats, error) {
	stmt := `
	SELECT user_id,
	COUNT (*) FILTER (WHERE created_at >= @start_date AND created_at <= @end_date) AS created_count,
	COUNT (*) FILTER (WHERE status = 'completed' AND completed_at >= @start_date AND completed_at <= @end_date) AS completed_count,
	COUNT (*) FILTER (WHERE status NOT IN('completed', 'archived')) AS active_count,
	COUNT (*) FILTER (WHERE due_date < NOW() AND status NOT IN ('completed', 'archived')) AS overdue_count
	FROM todos
	GROUP BY user_id
	HAVING COUNT (*) > 0
	`

	rows, err := r.db.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"start_date": startDate,
		"end_date":   endDate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get weekly stats for users query: %w", err)
	}

	stats, err := pgx.CollectRows(rows, pgx.RowToStructByName[todo.UserWeeklyStats])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []todo.UserWeeklyStats{}, nil
		}
		return nil, fmt.Errorf("failed to collect weekly stats for users: %w", err)
	}

	return stats, nil
}

func (r *TodoRepository) GetCompletedTodosForUser(ctx context.Context, userID string, startDate, endDate time.Time) ([]todo.PopulatedTodo, error) {
	stmt := populatedTodoSelectColumns + populatedTodoJoins + `
	WHERE t.user_id = @user_id
		AND t.status = 'completed'
		AND t.completed_at >= @start_date
		AND t.completed_at <= @end_date
	ORDER BY t.completed_at DESC
	LIMIT 10
	`

	rows, err := r.db.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"user_id":    userID,
		"start_date": startDate,
		"end_date":   endDate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get completed todos for user query: %w", err)
	}

	todos, err := pgx.CollectRows(rows, pgx.RowToStructByName[todo.PopulatedTodo])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []todo.PopulatedTodo{}, nil
		}
		return nil, fmt.Errorf("failed to collect completed todos for user: %w", err)
	}

	return todos, nil
}

func (r *TodoRepository) GetOverdueTodosForUser(ctx context.Context, userID string) ([]todo.PopulatedTodo, error) {
	stmt := populatedTodoSelectColumns + populatedTodoJoins + `
	WHERE t.user_id = @user_id
		AND t.status NOT IN ('completed', 'archived')
		AND t.due_date < NOW()
	ORDER BY t.due_date ASC
	LIMIT 10
	`

	rows, err := r.db.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"user_id": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get overdue todos for user query: %w", err)
	}

	todos, err := pgx.CollectRows(rows, pgx.RowToStructByName[todo.PopulatedTodo])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []todo.PopulatedTodo{}, nil
		}
		return nil, fmt.Errorf("failed to collect overdue todos for user: %w", err)
	}

	return todos, nil
}
