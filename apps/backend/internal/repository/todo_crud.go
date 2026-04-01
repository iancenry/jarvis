package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/iancenry/jarvis/internal/model"
	"github.com/iancenry/jarvis/internal/model/todo"
	"github.com/jackc/pgx/v5"
)

func (r *TodoRepository) CreateTodo(ctx context.Context, userID string, payload *todo.CreateTodoPayload) (*todo.Todo, error) {
	stmt := `
		INSERT INTO todos (
			user_id,
			title,
			description,
			priority,
			due_date,
			parent_todo_id,
			category_id,
			metadata
		)
		VALUES (
			@user_id,
			@title,
			@description,
			@priority,
			@due_date,
			@parent_todo_id,
			@category_id,
			@metadata
		)
		RETURNING *
	`

	priority := todo.PriorityMedium
	if payload.Priority != nil {
		priority = *payload.Priority
	}

	rows, err := r.db.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"user_id":        userID,
		"title":          payload.Title,
		"description":    payload.Description,
		"priority":       priority,
		"due_date":       payload.DueDate,
		"parent_todo_id": payload.ParentTodoID,
		"category_id":    payload.CategoryID,
		"metadata":       payload.Metadata,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute create todo query for user_id=%s title=%s: %w", userID, payload.Title, err)
	}

	todoItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[todo.Todo])
	if err != nil {
		return nil, fmt.Errorf("failed to collect created todo for user_id=%s title=%s: %w", userID, payload.Title, err)
	}

	return &todoItem, nil
}

func (r *TodoRepository) GetTodoByID(ctx context.Context, userID string, todoID uuid.UUID) (*todo.PopulatedTodo, error) {
	stmt := populatedTodoSelectColumns + populatedTodoJoins + `
		WHERE t.id = @todo_id AND t.user_id = @user_id
	`

	rows, err := r.db.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"todo_id": todoID,
		"user_id": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get todo by id query for user_id=%s todo_id=%s: %w", userID, todoID, err)
	}

	todoItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[todo.PopulatedTodo])
	if err != nil {
		if isNoRowsError(err) {
			return nil, newDomainNotFoundError("TODO")
		}
		return nil, fmt.Errorf("failed to collect todo by id for user_id=%s todo_id=%s: %w", userID, todoID, err)
	}

	return &todoItem, nil
}

func (r *TodoRepository) CheckTodoExists(ctx context.Context, userID string, todoID uuid.UUID) (*todo.Todo, error) {
	stmt := `
		SELECT * FROM todos
		WHERE id = @id AND user_id = @user_id
	`

	rows, err := r.db.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id":      todoID,
		"user_id": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check if todo exists query for user_id=%s id=%s: %w", userID, todoID, err)
	}

	todoItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[todo.Todo])
	if err != nil {
		if isNoRowsError(err) {
			return nil, newDomainNotFoundError("TODO")
		}
		return nil, fmt.Errorf("failed to collect todo exists for user_id=%s id=%s: %w", userID, todoID, err)
	}

	return &todoItem, nil
}

func (r *TodoRepository) GetTodos(ctx context.Context, userID string, query *todo.GetTodosQuery) (*model.PaginatedResponse[todo.PopulatedTodo], error) {
	stmt := populatedTodoSelectColumns + populatedTodoJoins

	args := pgx.NamedArgs{
		"user_id": userID,
	}
	conditions := []string{"t.user_id = @user_id"}

	if query.Status != nil {
		conditions = append(conditions, "t.status = @status")
		args["status"] = *query.Status
	}

	if query.Priority != nil {
		conditions = append(conditions, "t.priority = @priority")
		args["priority"] = *query.Priority
	}

	if query.CategoryID != nil {
		conditions = append(conditions, "t.category_id = @category_id")
		args["category_id"] = *query.CategoryID
	}

	if query.ParentTodoID != nil {
		conditions = append(conditions, "t.parent_todo_id = @parent_todo_id")
		args["parent_todo_id"] = *query.ParentTodoID
	} else {
		conditions = append(conditions, "t.parent_todo_id IS NULL")
	}

	if query.DueFrom != nil {
		conditions = append(conditions, "t.due_date >= @due_from")
		args["due_from"] = *query.DueFrom
	}

	if query.DueTo != nil {
		conditions = append(conditions, "t.due_date <= @due_to")
		args["due_to"] = *query.DueTo
	}

	if query.Overdue != nil && *query.Overdue {
		conditions = append(conditions, "t.due_date < NOW() AND t.status != 'completed'")
	}

	if query.Completed != nil {
		if *query.Completed {
			conditions = append(conditions, "t.status = 'completed'")
		} else {
			conditions = append(conditions, "t.status != 'completed'")
		}
	}

	if query.Search != nil {
		conditions = append(conditions, "(t.title ILIKE @search OR t.description ILIKE @search)")
		args["search"] = "%" + *query.Search + "%"
	}

	if len(conditions) > 0 {
		stmt += " WHERE " + strings.Join(conditions, " AND ")
	}

	countStmt := "SELECT COUNT(*) FROM todos t"
	if len(conditions) > 0 {
		countStmt += " WHERE " + strings.Join(conditions, " AND ")
	}

	var totalCount int
	err := r.db.Pool.QueryRow(ctx, countStmt, args).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to execute count todos query for user_id=%s: %w", userID, err)
	}

	sortColumn := todoSortColumns["created_at"]
	if query.Sort != nil {
		if mappedColumn, ok := todoSortColumns[*query.Sort]; ok {
			sortColumn = mappedColumn
		}
	}

	sortOrder := "DESC"
	if query.Order != nil && strings.EqualFold(*query.Order, "asc") {
		sortOrder = "ASC"
	}

	stmt += fmt.Sprintf(" ORDER BY %s %s", sortColumn, sortOrder)
	stmt += " LIMIT @limit OFFSET @offset"
	args["limit"] = *query.Limit
	args["offset"] = (*query.Page - 1) * (*query.Limit)

	rows, err := r.db.Pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get todos query for user_id=%s: %w", userID, err)
	}

	todos, err := pgx.CollectRows(rows, pgx.RowToStructByName[todo.PopulatedTodo])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &model.PaginatedResponse[todo.PopulatedTodo]{
				Data:       []todo.PopulatedTodo{},
				Page:       *query.Page,
				Limit:      *query.Limit,
				Total:      0,
				TotalPages: 0,
			}, nil
		}
		return nil, fmt.Errorf("failed to collect todos for user_id=%s: %w", userID, err)
	}

	return &model.PaginatedResponse[todo.PopulatedTodo]{
		Data:       todos,
		Page:       *query.Page,
		Limit:      *query.Limit,
		Total:      totalCount,
		TotalPages: (totalCount + *query.Limit - 1) / *query.Limit,
	}, nil
}

func (r *TodoRepository) UpdateTodo(ctx context.Context, userID string, payload *todo.UpdateTodoPayload) (*todo.Todo, error) {
	stmt := `UPDATE todos SET `
	args := pgx.NamedArgs{
		"id":      payload.ID,
		"user_id": userID,
	}

	setClauses := []string{}
	if payload.Title.IsSet() {
		if payload.Title.IsNull() {
			return nil, fmt.Errorf("title cannot be null for user_id=%s todo_id=%s", userID, payload.ID)
		}
		setClauses = append(setClauses, "title = @title")
		args["title"] = payload.Title.Value()
	}
	if payload.Description.IsSet() {
		setClauses = append(setClauses, "description = @description")
		if payload.Description.IsNull() {
			args["description"] = nil
		} else {
			args["description"] = payload.Description.Value()
		}
	}
	if payload.Status.IsSet() {
		if payload.Status.IsNull() {
			return nil, fmt.Errorf("status cannot be null for user_id=%s todo_id=%s", userID, payload.ID)
		}
		setClauses = append(setClauses, "status = @status")
		args["status"] = payload.Status.Value()

		if payload.Status.Value() == todo.StatusCompleted {
			setClauses = append(setClauses, "completed_at = @completed_at")
			args["completed_at"] = time.Now()
		} else {
			setClauses = append(setClauses, "completed_at = NULL")
		}
	}
	if payload.Priority.IsSet() {
		if payload.Priority.IsNull() {
			return nil, fmt.Errorf("priority cannot be null for user_id=%s todo_id=%s", userID, payload.ID)
		}
		setClauses = append(setClauses, "priority = @priority")
		args["priority"] = payload.Priority.Value()
	}
	if payload.DueDate.IsSet() {
		setClauses = append(setClauses, "due_date = @due_date")
		if payload.DueDate.IsNull() {
			args["due_date"] = nil
		} else {
			args["due_date"] = payload.DueDate.Value()
		}
	}
	if payload.ParentTodoID.IsSet() {
		setClauses = append(setClauses, "parent_todo_id = @parent_todo_id")
		if payload.ParentTodoID.IsNull() {
			args["parent_todo_id"] = nil
		} else {
			args["parent_todo_id"] = payload.ParentTodoID.Value()
		}
	}
	if payload.CategoryID.IsSet() {
		setClauses = append(setClauses, "category_id = @category_id")
		if payload.CategoryID.IsNull() {
			args["category_id"] = nil
		} else {
			args["category_id"] = payload.CategoryID.Value()
		}
	}
	if payload.Metadata.IsSet() {
		setClauses = append(setClauses, "metadata = @metadata")
		if payload.Metadata.IsNull() {
			args["metadata"] = nil
		} else {
			args["metadata"] = payload.Metadata.Value()
		}
	}

	if len(setClauses) == 0 {
		return nil, fmt.Errorf("no fields to update for user_id=%s todo_id=%s", userID, payload.ID)
	}

	stmt += strings.Join(setClauses, ", ")
	stmt += " WHERE id = @id AND user_id = @user_id RETURNING *"

	rows, err := r.db.Pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute update todo query for user_id=%s todo_id=%s: %w", userID, payload.ID, err)
	}

	todoItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[todo.Todo])
	if err != nil {
		if isNoRowsError(err) {
			return nil, newDomainNotFoundError("TODO")
		}
		return nil, fmt.Errorf("failed to collect updated todo for user_id=%s todo_id=%s: %w", userID, payload.ID, err)
	}

	return &todoItem, nil
}

func (r *TodoRepository) DeleteTodo(ctx context.Context, userID string, todoID uuid.UUID) (*todo.Todo, error) {
	stmt := `DELETE FROM todos WHERE id = @id AND user_id = @user_id RETURNING *`

	rows, err := r.db.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id":      todoID,
		"user_id": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute delete todo query for user_id=%s todo_id=%s: %w", userID, todoID, err)
	}

	deletedTodo, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[todo.Todo])
	if err != nil {
		if isNoRowsError(err) {
			return nil, newDomainNotFoundError("TODO")
		}
		return nil, fmt.Errorf("failed to collect deleted todo for user_id=%s todo_id=%s: %w", userID, todoID, err)
	}

	return &deletedTodo, nil
}

func (r *TodoRepository) GetTodoStats(ctx context.Context, userID string) (*todo.TodoStats, error) {
	stmt := `
	SELECT COUNT(*) AS total,
	COUNT (CASE WHEN status = 'draft' THEN 1 END) AS draft,
	COUNT (CASE WHEN status = 'active' THEN 1 END) AS active,
	COUNT (CASE WHEN status = 'completed' THEN 1 END) AS completed,
	COUNT (CASE WHEN status = 'archived' THEN 1 END) AS archived,
	COUNT (CASE WHEN due_date < NOW() AND status != 'completed' THEN 1 END) AS overdue
	FROM todos
	WHERE user_id = @user_id
	`

	rows, err := r.db.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"user_id": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get todo stats query for user_id=%s: %w", userID, err)
	}

	stats, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[todo.TodoStats])
	if err != nil {
		return nil, fmt.Errorf("failed to collect todo stats for user_id=%s: %w", userID, err)
	}

	return &stats, nil
}
