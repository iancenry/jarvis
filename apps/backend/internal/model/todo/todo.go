package todo

import (
	"time"

	"github.com/google/uuid"
	"github.com/iancenry/jarvis/internal/model"
	"github.com/iancenry/jarvis/internal/model/category"
	"github.com/iancenry/jarvis/internal/model/comment"
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
	StatusArchived  Status = "archived"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

//  TODO consider using a separate table for tags and a many-to-many relationship with todos
type Metadata struct{
	Tags []string `json:"tags"` 
	Reminder *string `json:"reminder,omitempty"`
	Color *string `json:"color,omitempty"`
	Difficulty *string `json:"difficulty,omitempty"`
}

// pointer type for optional fields; nullable in the database - else value types
type Todo struct {
	model.Base
	UserID      string   `json:"userId" db:"user_id"`
	Title       string   `json:"title" db:"title"`
	Description string   `json:"description" db:"description"`
	Status      Status   `json:"status" db:"status"`
	Priority    Priority `json:"priority" db:"priority"`
	DueDate 	*time.Time `json:"dueDate,omitempty" db:"due_date"`
	CompletedAt *time.Time `json:"completedAt,omitempty" db:"completed_at"`
	ParentTodoID *uuid.UUID `json:"parentTodoId,omitempty" db:"parent_todo_id"`
	CategoryID *uuid.UUID `json:"categoryId,omitempty" db:"category_id"`
	Metadata *Metadata `json:"metadata,omitempty" db:"metadata"`
	SortOrder int `json:"sortOrder" db:"sort_order"`
}

type PopulatedTodo struct {
	Todo
	Category *category.Category `json:"category,omitempty" db:"category,omitempty"`
	Children []Todo `json:"children,omitempty" db:"children,omitempty"`
	Comments []comment.Comment `json:"comments,omitempty" db:"comments,omitempty"`
}

type TodoStats struct {
	Total int `json:"total"`
	Draft int `json:"draft"`
	Active int `json:"active"`
	Completed int `json:"completed"`
	Archived int `json:"archived"`
	Overdue int `json:"overdue"`
}