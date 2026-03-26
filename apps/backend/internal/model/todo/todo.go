package todo

import (
	"time"

	"github.com/google/uuid"
	"github.com/iancenry/jarvis/internal/model"
	"github.com/iancenry/jarvis/internal/model/attachment"
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
	Category *category.Category `json:"category" db:"category"`
	Children []Todo `json:"children" db:"children"`
	Comments []comment.Comment `json:"comments" db:"comments"`
	Attachments []attachment.Attachment `json:"attachments" db:"attachments"`
}

type TodoStats struct {
	Total int `json:"total"`
	Draft int `json:"draft"`
	Active int `json:"active"`
	Completed int `json:"completed"`
	Archived int `json:"archived"`
	Overdue int `json:"overdue"`
}

type UserWeeklyStats struct {
	UserID string `json:"userId" db:"user_id"`
	CreatedCount int `json:"createdCount" db:"created_count"`
	CompletedCount int `json:"completedCount" db:"completed_count"`
	ActiveCount int `json:"activeCount" db:"active_count"`
	OverdueCount int `json:"overdueCount" db:"overdue_count"`
}

// subtasks can't have subtasks
func (t *Todo) CanHaveChildren() bool {
	return t.ParentTodoID == nil
}

func (t *Todo) IsOverdue() bool {
	return t.DueDate != nil && t.DueDate.Before(time.Now()) && t.Status != StatusCompleted
}