package repository

import "github.com/iancenry/jarvis/internal/database"

// Repositories aggregates all individual repositories for easy access
type Repositories struct {
	Todo     *TodoRepository
	Comment  *CommentRepository
	Category *CategoryRepository
}

func NewRepositories(db *database.Database) *Repositories {
	return &Repositories{
		Todo:     NewTodoRepository(db),
		Comment:  NewCommentRepository(db),
		Category: NewCategoryRepository(db),
	}
}
