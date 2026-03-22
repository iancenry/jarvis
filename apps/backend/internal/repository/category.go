package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/iancenry/jarvis/internal/model"
	"github.com/iancenry/jarvis/internal/model/category"
	"github.com/iancenry/jarvis/internal/server"
	"github.com/jackc/pgx/v5"
)

type CategoryRepository struct {
	server *server.Server
}

func NewCategoryRepository(s *server.Server) *CategoryRepository {
	return &CategoryRepository{
		server: s,
	}
}

func (r *CategoryRepository) CreateCategory(ctx context.Context, userID string, payload *category.CreateCategoryPayload) (*category.Category, error) {
	stmt := `
		INSERT INTO categories (user_id, name, color, description)
		VALUES (@user_id, @name, @color, @description)
		RETURNING *
	`
	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"user_id": userID,
		"name": payload.Name,
		"color": payload.Color,
		"description": payload.Description,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute create category query for user_id=%s: %w", userID, err)
	}

	category, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[category.Category])
	if err != nil {
		return nil, fmt.Errorf("failed to collect created category for user_id=%s: %w", userID, err)
	}

	return &category, nil
}

func (r *CategoryRepository) GetCategories(ctx context.Context, userID string, query *category.GetCategoriesQuery) (*model.PaginatedResponse[category.Category], error) {
	stmt := `
		SELECT *
		FROM categories
		WHERE user_id = @user_id
	`

	args := pgx.NamedArgs{
		"user_id": userID,
	}

	if query.Search != nil {
		stmt += " AND name ILIKE '%' || @search || '%'"
		args["search"] = *query.Search
	}

	sortColumn := "name"
	if query.Sort != nil {
		sortColumn = *query.Sort
	}
	sortOrder := "asc"
	if query.Order != nil {
		sortOrder = *query.Order
	}

	stmt += fmt.Sprintf(" ORDER BY %s %s", sortColumn, sortOrder)
	stmt += " LIMIT @limit OFFSET @offset"
	args["limit"] = *query.Limit
	args["offset"] = (*query.Page - 1) * *query.Limit

	rows, err := r.server.DB.Pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get categories query for user_id=%s: %w", userID, err)
	}

	categories, err := pgx.CollectRows(rows, pgx.RowToStructByName[category.Category])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &model.PaginatedResponse[category.Category]{
				Data: []category.Category{},
				Page: *query.Page,
				Limit: *query.Limit,
				Total: 0,
				TotalPages: 0,
			}, nil
		}

		return nil, fmt.Errorf("failed to collect categories for user_id=%s: %w", userID, err)
	}

	countStmt := `
		SELECT COUNT(*)
		FROM categories
		WHERE user_id = @user_id
	`
	if query.Search != nil {
		countStmt += " AND name ILIKE '%' || @search || '%'"
	}
	var total int
	err = r.server.DB.Pool.QueryRow(ctx, countStmt, args).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to execute count categories query for user_id=%s: %w", userID, err)
	}

	return &model.PaginatedResponse[category.Category]{
		Data: categories,
		Page: *query.Page,
		Limit: *query.Limit,
		Total: total,
		TotalPages: (total + *query.Limit - 1) / *query.Limit,
	}, nil

}

func (r *CategoryRepository) UpdateCategory(ctx context.Context, userID string, categoryID string, payload *category.UpdateCategoryPayload) (*category.Category, error) {
	stmt := `UPDATE todo_categories SET `

	args := pgx.NamedArgs{
		"user_id": userID,
		"id": categoryID,
	}

	setClauses := []string{}
	if payload.Name != nil {
		setClauses = append(setClauses, "name = @name")
		args["name"] = *payload.Name
	}
	if payload.Color != nil {
		setClauses = append(setClauses, "color = @color")
		args["color"] = *payload.Color
	}
	if payload.Description != nil {
		setClauses = append(setClauses, "description = @description")
		args["description"] = *payload.Description
	}

	if len(setClauses) == 0 {
		return nil, fmt.Errorf("no fields to update for category_id=%s", categoryID)
	}

	stmt += strings.Join(setClauses, ", ")
	stmt += " WHERE user_id = @user_id AND id = @id RETURNING *"

	rows, err := r.server.DB.Pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute update category query for user_id=%s and category_id=%s: %w", userID, categoryID, err)
	}

	updatedCategory, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[category.Category])
	if err != nil {
		return nil, fmt.Errorf("failed to collect updated category for user_id=%s and category_id=%s: %w", userID, categoryID, err)
	}

	return &updatedCategory, nil
}


func (r *CategoryRepository) DeleteCategory(ctx context.Context, userID string, categoryID string) error {
	stmt := `
		DELETE FROM categories
		WHERE user_id = @user_id AND id = @id
	`
	result, err := r.server.DB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"user_id": userID,
		"id": categoryID,
	})
	if err != nil {
		return fmt.Errorf("failed to execute delete category query for user_id=%s and category_id=%s: %w", userID, categoryID, err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no category found to delete for user_id=%s and category_id=%s", userID, categoryID)
	}

	return nil
}