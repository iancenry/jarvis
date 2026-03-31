package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/iancenry/jarvis/internal/database"
	"github.com/iancenry/jarvis/internal/model"
	"github.com/iancenry/jarvis/internal/model/category"
	"github.com/jackc/pgx/v5"
)

type CategoryRepository struct {
	db *database.Database
}

func NewCategoryRepository(db *database.Database) *CategoryRepository {
	return &CategoryRepository{
		db: db,
	}
}

func (r *CategoryRepository) CreateCategory(ctx context.Context, userID string, payload *category.CreateCategoryPayload) (*category.Category, error) {
	stmt := `
		INSERT INTO todo_categories (user_id, name, color, description)
		VALUES (@user_id, @name, @color, @description)
		RETURNING *
	`
	rows, err := r.db.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"user_id":     userID,
		"name":        payload.Name,
		"color":       payload.Color,
		"description": payload.Description,
	})

	if err != nil {
		if isUniqueViolation(err) {
			return nil, newDomainConflictError("CATEGORY", "category with this name already exists")
		}
		return nil, fmt.Errorf("failed to execute create category query for user_id=%s: %w", userID, err)
	}

	category, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[category.Category])
	if err != nil {
		if isUniqueViolation(err) {
			return nil, newDomainConflictError("CATEGORY", "category with this name already exists")
		}
		return nil, fmt.Errorf("failed to collect created category for user_id=%s: %w", userID, err)
	}

	return &category, nil
}

func (r *CategoryRepository) GetCategories(ctx context.Context, userID string, query *category.GetCategoriesQuery) (*model.PaginatedResponse[category.Category], error) {
	stmt := `
		SELECT *
		FROM todo_categories
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

	rows, err := r.db.Pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get categories query for user_id=%s: %w", userID, err)
	}

	categories, err := pgx.CollectRows(rows, pgx.RowToStructByName[category.Category])
	if err != nil {
		return nil, fmt.Errorf("failed to collect categories for user_id=%s: %w", userID, err)
	}

	countStmt := `
		SELECT COUNT(*)
		FROM todo_categories
		WHERE user_id = @user_id
	`
	if query.Search != nil {
		countStmt += " AND name ILIKE '%' || @search || '%'"
	}
	var total int
	err = r.db.Pool.QueryRow(ctx, countStmt, args).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to execute count categories query for user_id=%s: %w", userID, err)
	}

	return &model.PaginatedResponse[category.Category]{
		Data:       categories,
		Page:       *query.Page,
		Limit:      *query.Limit,
		Total:      total,
		TotalPages: (total + *query.Limit - 1) / *query.Limit,
	}, nil

}

func (r *CategoryRepository) GetCategoryByID(ctx context.Context, userID string, categoryID uuid.UUID) (*category.Category, error) {
	stmt := `
		SELECT *
		FROM todo_categories
		WHERE user_id = @user_id AND id = @id
	`
	rows, err := r.db.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"user_id": userID,
		"id":      categoryID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get category by ID query for user_id=%s and category_id=%s: %w", userID, categoryID, err)
	}

	category, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[category.Category])
	if err != nil {
		if isNoRowsError(err) {
			return nil, newDomainNotFoundError("CATEGORY")
		}
		return nil, fmt.Errorf("failed to collect category by ID for user_id=%s and category_id=%s: %w", userID, categoryID, err)
	}

	return &category, nil
}

func (r *CategoryRepository) UpdateCategory(ctx context.Context, userID string, categoryID uuid.UUID, payload *category.UpdateCategoryPayload) (*category.Category, error) {
	stmt := `UPDATE todo_categories SET `

	args := pgx.NamedArgs{
		"user_id": userID,
		"id":      categoryID,
	}

	setClauses := []string{}
	if payload.Name.IsSet() {
		if payload.Name.IsNull() {
			return nil, fmt.Errorf("name cannot be null for category_id=%s", categoryID)
		}
		setClauses = append(setClauses, "name = @name")
		args["name"] = payload.Name.Value()
	}
	if payload.Color.IsSet() {
		if payload.Color.IsNull() {
			return nil, fmt.Errorf("color cannot be null for category_id=%s", categoryID)
		}
		setClauses = append(setClauses, "color = @color")
		args["color"] = payload.Color.Value()
	}
	if payload.Description.IsSet() {
		setClauses = append(setClauses, "description = @description")
		if payload.Description.IsNull() {
			args["description"] = nil
		} else {
			args["description"] = payload.Description.Value()
		}
	}

	if len(setClauses) == 0 {
		return nil, fmt.Errorf("no fields to update for category_id=%s", categoryID)
	}

	stmt += strings.Join(setClauses, ", ")
	stmt += " WHERE user_id = @user_id AND id = @id RETURNING *"

	rows, err := r.db.Pool.Query(ctx, stmt, args)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, newDomainConflictError("CATEGORY", "category with this name already exists")
		}
		return nil, fmt.Errorf("failed to execute update category query for user_id=%s and category_id=%s: %w", userID, categoryID, err)
	}

	updatedCategory, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[category.Category])
	if err != nil {
		if isUniqueViolation(err) {
			return nil, newDomainConflictError("CATEGORY", "category with this name already exists")
		}
		if isNoRowsError(err) {
			return nil, newDomainNotFoundError("CATEGORY")
		}
		return nil, fmt.Errorf("failed to collect updated category for user_id=%s and category_id=%s: %w", userID, categoryID, err)
	}

	return &updatedCategory, nil
}

func (r *CategoryRepository) DeleteCategory(ctx context.Context, userID string, categoryID uuid.UUID) (*category.Category, error) {
	stmt := `
		DELETE FROM todo_categories
		WHERE user_id = @user_id AND id = @id
		RETURNING *
	`
	rows, err := r.db.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"user_id": userID,
		"id":      categoryID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute delete category query for user_id=%s and category_id=%s: %w", userID, categoryID, err)
	}

	deletedCategory, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[category.Category])
	if err != nil {
		if isNoRowsError(err) {
			return nil, newDomainNotFoundError("CATEGORY")
		}
		return nil, fmt.Errorf("failed to collect deleted category for user_id=%s and category_id=%s: %w", userID, categoryID, err)
	}

	return &deletedCategory, nil
}
