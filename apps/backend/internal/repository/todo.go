package repository

import (
	"github.com/iancenry/jarvis/internal/database"
)

// TodoRepository provides methods to interact with the todo data store
type TodoRepository struct {
	db *database.Database
}

// todoSortColumns maps allowed sort fields from the API to their corresponding database columns to prevent SQL injection through the sort parameter
var todoSortColumns = map[string]string{
	"created_at": "t.created_at",
	"updated_at": "t.updated_at",
	"due_date":   "t.due_date",
	"priority":   "t.priority",
}

const populatedTodoSelectColumns = `
	SELECT t.*,
	CASE
		WHEN c.id IS NOT NULL THEN jsonb(camel (c))
		ELSE NULL
	END AS category,
	COALESCE(child_agg.children, '[]'::jsonb) AS children,
	COALESCE(comment_agg.comments, '[]'::jsonb) AS comments,
	COALESCE(attachment_agg.attachments, '[]'::jsonb) AS attachments
`

const populatedTodoJoins = `
	FROM todos t
		LEFT JOIN todo_categories c ON t.category_id = c.id AND c.user_id = @user_id
		LEFT JOIN LATERAL (
			SELECT jsonb_agg(
				to_jsonb(camel (child))
				ORDER BY child.created_at ASC, child.sort_order ASC
			) AS children
			FROM todos child
			WHERE child.parent_todo_id = t.id AND child.user_id = @user_id
		) child_agg ON TRUE
		LEFT JOIN LATERAL (
			SELECT jsonb_agg(
				to_jsonb(camel (com))
				ORDER BY com.created_at ASC
			) AS comments
			FROM todo_comments com
			WHERE com.todo_id = t.id AND com.user_id = @user_id
		) comment_agg ON TRUE
		LEFT JOIN LATERAL (
			SELECT jsonb_agg(
				to_jsonb(camel (att))
				ORDER BY att.created_at ASC
			) AS attachments
			FROM attachments att
			WHERE att.todo_id = t.id
		) attachment_agg ON TRUE
`

// NewTodoRepository creates a new instance of TodoRepository with the provided database dependency.
func NewTodoRepository(db *database.Database) *TodoRepository {
	return &TodoRepository{
		db: db,
	}
}
