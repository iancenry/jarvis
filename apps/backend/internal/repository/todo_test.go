package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	attachmentModel "github.com/iancenry/jarvis/internal/model/attachment"
	commentModel "github.com/iancenry/jarvis/internal/model/comment"
	"github.com/iancenry/jarvis/internal/model/todo"
	"github.com/iancenry/jarvis/internal/repository"
	testing_pkg "github.com/iancenry/jarvis/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTodoRepository_CreateTodo(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	todoRepo := repository.NewTodoRepository(testServer)

	t.Run("create todo successfully", func(t *testing.T) {
		userID := uuid.New().String()
		dueDate := time.Now().Add(24 * time.Hour)
		payload := &todo.CreateTodoPayload{
			Title:       "Test Todo",
			Description: testing_pkg.Ptr("Test todo description"),
			Priority:    testing_pkg.Ptr(todo.PriorityHigh),
			DueDate:     &dueDate,
		}

		result, err := todoRepo.CreateTodo(ctx, userID, payload)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.NotEqual(t, uuid.Nil, result.ID)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, payload.Title, result.Title)
		assert.Equal(t, payload.Description, result.Description)
		assert.Equal(t, *payload.Priority, result.Priority)
		assert.Equal(t, payload.DueDate.Unix(), result.DueDate.Unix())
		assert.Equal(t, todo.StatusDraft, result.Status)
		assert.Nil(t, result.CompletedAt)
		testing_pkg.AssertTimestampsValid(t, result)
	})

	t.Run("create todo with minimum required fields", func(t *testing.T) {
		userID := uuid.New().String()
		payload := &todo.CreateTodoPayload{
			Title: "Minimal Todo",
		}

		result, err := todoRepo.CreateTodo(ctx, userID, payload)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, payload.Title, result.Title)
		assert.Nil(t, result.Description)
		assert.Equal(t, todo.PriorityMedium, result.Priority)
		assert.Nil(t, result.DueDate)
	})

	t.Run("create todo with metadata", func(t *testing.T) {
		userID := uuid.New().String()
		metadata := &todo.Metadata{
			Tags:  []string{"work", "urgent"},
			Color: testing_pkg.Ptr("#ff0000"),
		}
		payload := &todo.CreateTodoPayload{
			Title:    "Todo with Metadata",
			Metadata: metadata,
		}

		result, err := todoRepo.CreateTodo(ctx, userID, payload)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, metadata.Tags, result.Metadata.Tags)
		assert.Equal(t, metadata.Color, result.Metadata.Color)
	})

	t.Run("with canceled context", func(t *testing.T) {
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		userID := uuid.New().String()
		payload := &todo.CreateTodoPayload{
			Title: "Canceled Todo",
		}

		result, err := todoRepo.CreateTodo(canceledCtx, userID, payload)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestTodoRepository_GetTodoByID(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	todoRepo := repository.NewTodoRepository(testServer)

	// Create test todo
	userID := uuid.New().String()
	testTodo := createTestTodo(t, ctx, todoRepo, userID)

	t.Run("get todo by id successfully", func(t *testing.T) {
		result, err := todoRepo.GetTodoByID(ctx, userID, testTodo.ID)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, testTodo.ID, result.ID)
		assert.Equal(t, testTodo.Title, result.Title)
		assert.Equal(t, testTodo.UserID, result.UserID)
		assert.NotNil(t, result.Children)
		assert.NotNil(t, result.Comments)
	})

	t.Run("does not duplicate nested relations", func(t *testing.T) {
		commentRepo := repository.NewCommentRepository(testServer)
		fixture := createPopulatedTodoFixture(t, ctx, todoRepo, commentRepo, userID, "Nested GetTodoByID", time.Now().Add(6*time.Hour))

		result, err := todoRepo.GetTodoByID(ctx, userID, fixture.Parent.ID)
		require.NoError(t, err)
		require.NotNil(t, result)

		assertPopulatedTodoFixture(t, result, fixture)
	})

	t.Run("get non-existent todo", func(t *testing.T) {
		nonExistentID := uuid.New()

		result, err := todoRepo.GetTodoByID(ctx, userID, nonExistentID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("get todo with wrong user id", func(t *testing.T) {
		wrongUserID := uuid.New().String()

		result, err := todoRepo.GetTodoByID(ctx, wrongUserID, testTodo.ID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("with canceled context", func(t *testing.T) {
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		result, err := todoRepo.GetTodoByID(canceledCtx, userID, testTodo.ID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestTodoRepository_CheckTodoExists(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	todoRepo := repository.NewTodoRepository(testServer)

	// Create test todo
	userID := uuid.New().String()
	testTodo := createTestTodo(t, ctx, todoRepo, userID)

	t.Run("check existing todo", func(t *testing.T) {
		result, err := todoRepo.CheckTodoExists(ctx, userID, testTodo.ID)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, testTodo.ID, result.ID)
		assert.Equal(t, testTodo.Title, result.Title)
	})

	t.Run("check non-existent todo", func(t *testing.T) {
		nonExistentID := uuid.New()

		result, err := todoRepo.CheckTodoExists(ctx, userID, nonExistentID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("check todo with wrong user id", func(t *testing.T) {
		wrongUserID := uuid.New().String()

		result, err := todoRepo.CheckTodoExists(ctx, wrongUserID, testTodo.ID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestTodoRepository_GetTodos(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	todoRepo := repository.NewTodoRepository(testServer)

	// Create test todos
	userID := uuid.New().String()
	_ = createTestTodos(t, ctx, todoRepo, userID, 3)

	t.Run("get todos with default pagination", func(t *testing.T) {
		page := 1
		limit := 20
		query := &todo.GetTodosQuery{
			Page:  &page,
			Limit: &limit,
		}

		result, err := todoRepo.GetTodos(ctx, userID, query)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.GreaterOrEqual(t, len(result.Data), 3)
		assert.Equal(t, page, result.Page)
		assert.Equal(t, limit, result.Limit)
		assert.GreaterOrEqual(t, result.Total, 3)
		assert.GreaterOrEqual(t, result.TotalPages, 1)
	})

	t.Run("get todos with pagination", func(t *testing.T) {
		page := 1
		limit := 2
		query := &todo.GetTodosQuery{
			Page:  &page,
			Limit: &limit,
		}

		result, err := todoRepo.GetTodos(ctx, userID, query)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, page, result.Page)
		assert.Equal(t, limit, result.Limit)
		assert.GreaterOrEqual(t, result.Total, 3)
		assert.GreaterOrEqual(t, result.TotalPages, 2)
	})

	t.Run("filter by status", func(t *testing.T) {
		page := 1
		limit := 20
		status := todo.StatusDraft
		query := &todo.GetTodosQuery{
			Page:   &page,
			Limit:  &limit,
			Status: &status,
		}

		result, err := todoRepo.GetTodos(ctx, userID, query)
		require.NoError(t, err)
		require.NotNil(t, result)

		for _, todoItem := range result.Data {
			assert.Equal(t, todo.StatusDraft, todoItem.Status)
		}
	})

	t.Run("filter by priority", func(t *testing.T) {
		page := 1
		limit := 20
		priority := todo.PriorityHigh
		query := &todo.GetTodosQuery{
			Page:     &page,
			Limit:    &limit,
			Priority: &priority,
		}

		result, err := todoRepo.GetTodos(ctx, userID, query)
		require.NoError(t, err)
		require.NotNil(t, result)

		for _, todoItem := range result.Data {
			assert.Equal(t, todo.PriorityHigh, todoItem.Priority)
		}
	})

	t.Run("search by title", func(t *testing.T) {
		page := 1
		limit := 20
		search := "Test"
		query := &todo.GetTodosQuery{
			Page:   &page,
			Limit:  &limit,
			Search: &search,
		}

		result, err := todoRepo.GetTodos(ctx, userID, query)
		require.NoError(t, err)
		require.NotNil(t, result)

		for _, todoItem := range result.Data {
			assert.Contains(t, todoItem.Title, "Test")
		}
	})

	t.Run("sort by due date ascending", func(t *testing.T) {
		page := 1
		limit := 20
		sort := "due_date"
		order := "asc"
		query := &todo.GetTodosQuery{
			Page:  &page,
			Limit: &limit,
			Sort:  &sort,
			Order: &order,
		}

		result, err := todoRepo.GetTodos(ctx, userID, query)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Data, 3)

		assert.Equal(t, "Test Todo 1", result.Data[0].Title)
		assert.Equal(t, "Test Todo 2", result.Data[1].Title)
		assert.Equal(t, "Test Todo 3", result.Data[2].Title)
	})

	t.Run("does not duplicate nested relations", func(t *testing.T) {
		commentRepo := repository.NewCommentRepository(testServer)
		nestedUserID := uuid.New().String()
		fixture := createPopulatedTodoFixture(t, ctx, todoRepo, commentRepo, nestedUserID, "Nested GetTodos", time.Now().Add(8*time.Hour))

		page := 1
		limit := 20
		query := &todo.GetTodosQuery{
			Page:  &page,
			Limit: &limit,
		}

		result, err := todoRepo.GetTodos(ctx, nestedUserID, query)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Data, 1)

		assertPopulatedTodoFixture(t, &result.Data[0], fixture)
	})

	t.Run("with canceled context", func(t *testing.T) {
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		page := 1
		limit := 20
		query := &todo.GetTodosQuery{
			Page:  &page,
			Limit: &limit,
		}

		result, err := todoRepo.GetTodos(canceledCtx, userID, query)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestTodoRepository_UpdateTodo(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	todoRepo := repository.NewTodoRepository(testServer)

	// Create test todo
	userID := uuid.New().String()
	testTodo := createTestTodo(t, ctx, todoRepo, userID)

	t.Run("update todo title successfully", func(t *testing.T) {
		newTitle := "Updated Todo Title"
		payload := &todo.UpdateTodoPayload{
			ID:    testTodo.ID,
			Title: &newTitle,
		}

		result, err := todoRepo.UpdateTodo(ctx, userID, payload)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, testTodo.ID, result.ID)
		assert.Equal(t, newTitle, result.Title)
		assert.True(t, result.UpdatedAt.After(testTodo.UpdatedAt))
	})

	t.Run("update todo status to completed", func(t *testing.T) {
		status := todo.StatusCompleted
		payload := &todo.UpdateTodoPayload{
			ID:     testTodo.ID,
			Status: &status,
		}

		result, err := todoRepo.UpdateTodo(ctx, userID, payload)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, todo.StatusCompleted, result.Status)
		assert.NotNil(t, result.CompletedAt)
	})

	t.Run("update multiple fields successfully", func(t *testing.T) {
		newTitle := "Multi Update Todo"
		newPriority := todo.PriorityLow
		payload := &todo.UpdateTodoPayload{
			ID:       testTodo.ID,
			Title:    &newTitle,
			Priority: &newPriority,
		}

		result, err := todoRepo.UpdateTodo(ctx, userID, payload)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, newTitle, result.Title)
		assert.Equal(t, newPriority, result.Priority)
	})

	t.Run("update with no fields should fail", func(t *testing.T) {
		payload := &todo.UpdateTodoPayload{
			ID: testTodo.ID,
		}

		result, err := todoRepo.UpdateTodo(ctx, userID, payload)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "no fields to update")
	})

	t.Run("update non-existent todo", func(t *testing.T) {
		nonExistentID := uuid.New()
		newTitle := "Non Existent Todo"
		payload := &todo.UpdateTodoPayload{
			ID:    nonExistentID,
			Title: &newTitle,
		}

		result, err := todoRepo.UpdateTodo(ctx, userID, payload)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("with canceled context", func(t *testing.T) {
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		newTitle := "Canceled Update"
		payload := &todo.UpdateTodoPayload{
			ID:    testTodo.ID,
			Title: &newTitle,
		}

		result, err := todoRepo.UpdateTodo(canceledCtx, userID, payload)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestTodoRepository_DeleteTodo(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	todoRepo := repository.NewTodoRepository(testServer)

	// Create test todo
	userID := uuid.New().String()
	testTodo := createTestTodo(t, ctx, todoRepo, userID)

	t.Run("delete todo successfully", func(t *testing.T) {
		err := todoRepo.DeleteTodo(ctx, userID, testTodo.ID)
		require.NoError(t, err)

		// Verify todo is deleted
		result, err := todoRepo.GetTodoByID(ctx, userID, testTodo.ID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("delete non-existent todo", func(t *testing.T) {
		nonExistentID := uuid.New()

		err := todoRepo.DeleteTodo(ctx, userID, nonExistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "todo not found")
	})

	t.Run("with canceled context", func(t *testing.T) {
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		testTodo := createTestTodo(t, ctx, todoRepo, userID)

		err := todoRepo.DeleteTodo(canceledCtx, userID, testTodo.ID)
		assert.Error(t, err)
	})
}

func TestTodoRepository_GetTodoStats(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	todoRepo := repository.NewTodoRepository(testServer)

	// Create test todos with different statuses
	userID := uuid.New().String()
	createTestTodos(t, ctx, todoRepo, userID, 5)

	t.Run("get todo stats successfully", func(t *testing.T) {
		result, err := todoRepo.GetTodoStats(ctx, userID)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.GreaterOrEqual(t, result.Total, 5)
		assert.GreaterOrEqual(t, result.Draft, 0)
		assert.GreaterOrEqual(t, result.Active, 0)
		assert.GreaterOrEqual(t, result.Completed, 0)
		assert.GreaterOrEqual(t, result.Archived, 0)
		assert.GreaterOrEqual(t, result.Overdue, 0)
	})

	t.Run("with canceled context", func(t *testing.T) {
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		result, err := todoRepo.GetTodoStats(canceledCtx, userID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestTodoRepository_GetTodosDueInHours(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	todoRepo := repository.NewTodoRepository(testServer)
	userID := uuid.New().String()

	dueSoon := createTestTodoWithDueDate(t, ctx, todoRepo, userID, "Due Soon", time.Now().Add(1*time.Hour))
	dueSoonLater := createTestTodoWithDueDate(t, ctx, todoRepo, userID, "Due Soon Later", time.Now().Add(2*time.Hour))
	_ = createTestTodoWithDueDate(t, ctx, todoRepo, userID, "Due Later", time.Now().Add(24*time.Hour))
	_ = createTestTodoWithDueDate(t, ctx, todoRepo, userID, "Already Overdue", time.Now().Add(-2*time.Hour))

	completedTodo := createTestTodoWithDueDate(t, ctx, todoRepo, userID, "Completed Soon", time.Now().Add(90*time.Minute))
	setTodoStatus(t, ctx, todoRepo, userID, completedTodo.ID, todo.StatusCompleted)

	archivedTodo := createTestTodoWithDueDate(t, ctx, todoRepo, userID, "Archived Soon", time.Now().Add(75*time.Minute))
	setTodoStatus(t, ctx, todoRepo, userID, archivedTodo.ID, todo.StatusArchived)

	result, err := todoRepo.GetTodosDueInHours(ctx, 3, 10)
	require.NoError(t, err)
	require.Len(t, result, 2)

	assert.Equal(t, dueSoon.ID, result[0].ID)
	assert.Equal(t, dueSoonLater.ID, result[1].ID)
	assert.Equal(t, "Due Soon", result[0].Title)
	assert.Equal(t, "Due Soon Later", result[1].Title)
}

func TestTodoRepository_GetOverdueTodos(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	todoRepo := repository.NewTodoRepository(testServer)
	userID := uuid.New().String()

	oldestOverdue := createTestTodoWithDueDate(t, ctx, todoRepo, userID, "Oldest Overdue", time.Now().Add(-4*time.Hour))
	newerOverdue := createTestTodoWithDueDate(t, ctx, todoRepo, userID, "Newer Overdue", time.Now().Add(-2*time.Hour))
	_ = createTestTodoWithDueDate(t, ctx, todoRepo, userID, "Future Todo", time.Now().Add(4*time.Hour))

	completedOverdue := createTestTodoWithDueDate(t, ctx, todoRepo, userID, "Completed Overdue", time.Now().Add(-90*time.Minute))
	setTodoStatus(t, ctx, todoRepo, userID, completedOverdue.ID, todo.StatusCompleted)

	archivedOverdue := createTestTodoWithDueDate(t, ctx, todoRepo, userID, "Archived Overdue", time.Now().Add(-75*time.Minute))
	setTodoStatus(t, ctx, todoRepo, userID, archivedOverdue.ID, todo.StatusArchived)

	result, err := todoRepo.GetOverdueTodos(ctx, 10)
	require.NoError(t, err)
	require.Len(t, result, 2)

	assert.Equal(t, oldestOverdue.ID, result[0].ID)
	assert.Equal(t, newerOverdue.ID, result[1].ID)
	assert.Equal(t, "Oldest Overdue", result[0].Title)
	assert.Equal(t, "Newer Overdue", result[1].Title)
}

func TestTodoRepository_AutoArchiveQueries(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	todoRepo := repository.NewTodoRepository(testServer)
	userID := uuid.New().String()

	oldCompletedA := createTestTodoWithDueDate(t, ctx, todoRepo, userID, "Old Completed A", time.Now().Add(-48*time.Hour))
	oldCompletedB := createTestTodoWithDueDate(t, ctx, todoRepo, userID, "Old Completed B", time.Now().Add(-72*time.Hour))
	recentCompleted := createTestTodoWithDueDate(t, ctx, todoRepo, userID, "Recent Completed", time.Now().Add(-24*time.Hour))
	_ = createTestTodoWithDueDate(t, ctx, todoRepo, userID, "Still Active", time.Now().Add(-96*time.Hour))

	setTodoStatus(t, ctx, todoRepo, userID, oldCompletedA.ID, todo.StatusCompleted)
	setTodoStatus(t, ctx, todoRepo, userID, oldCompletedB.ID, todo.StatusCompleted)
	setTodoStatus(t, ctx, todoRepo, userID, recentCompleted.ID, todo.StatusCompleted)

	oldCompletionTime := time.Now().AddDate(0, 0, -45)
	recentCompletionTime := time.Now().AddDate(0, 0, -5)

	_, err := testServer.DB.Pool.Exec(
		ctx,
		"UPDATE todos SET completed_at = $1 WHERE id = $2",
		oldCompletionTime,
		oldCompletedA.ID,
	)
	require.NoError(t, err)

	_, err = testServer.DB.Pool.Exec(
		ctx,
		"UPDATE todos SET completed_at = $1 WHERE id = $2",
		oldCompletionTime.Add(-time.Hour),
		oldCompletedB.ID,
	)
	require.NoError(t, err)

	_, err = testServer.DB.Pool.Exec(
		ctx,
		"UPDATE todos SET completed_at = $1 WHERE id = $2",
		recentCompletionTime,
		recentCompleted.ID,
	)
	require.NoError(t, err)

	cutoffDate := time.Now().AddDate(0, 0, -30)

	eligibleTodos, err := todoRepo.GetCompletedTodosOlderThan(ctx, cutoffDate, 10)
	require.NoError(t, err)
	require.Len(t, eligibleTodos, 2)
	assert.ElementsMatch(t, []uuid.UUID{oldCompletedA.ID, oldCompletedB.ID}, []uuid.UUID{eligibleTodos[0].ID, eligibleTodos[1].ID})

	err = todoRepo.ArchiveTodos(ctx, []uuid.UUID{oldCompletedA.ID, oldCompletedB.ID})
	require.NoError(t, err)

	archivedA, err := todoRepo.CheckTodoExists(ctx, userID, oldCompletedA.ID)
	require.NoError(t, err)
	assert.Equal(t, todo.StatusArchived, archivedA.Status)

	archivedB, err := todoRepo.CheckTodoExists(ctx, userID, oldCompletedB.ID)
	require.NoError(t, err)
	assert.Equal(t, todo.StatusArchived, archivedB.Status)

	stillCompleted, err := todoRepo.CheckTodoExists(ctx, userID, recentCompleted.ID)
	require.NoError(t, err)
	assert.Equal(t, todo.StatusCompleted, stillCompleted.Status)

	eligibleTodos, err = todoRepo.GetCompletedTodosOlderThan(ctx, cutoffDate, 10)
	require.NoError(t, err)
	assert.Empty(t, eligibleTodos)
}

func TestTodoRepository_ReportQueriesDoNotDuplicateNestedRelations(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	todoRepo := repository.NewTodoRepository(testServer)
	commentRepo := repository.NewCommentRepository(testServer)
	userID := uuid.New().String()

	completedFixture := createPopulatedTodoFixture(t, ctx, todoRepo, commentRepo, userID, "Completed Report", time.Now().Add(-4*time.Hour))
	completedTodo := setTodoStatus(t, ctx, todoRepo, userID, completedFixture.Parent.ID, todo.StatusCompleted)
	require.NotNil(t, completedTodo.CompletedAt)

	overdueFixture := createPopulatedTodoFixture(t, ctx, todoRepo, commentRepo, userID, "Overdue Report", time.Now().Add(-2*time.Hour))

	completedTodos, err := todoRepo.GetCompletedTodosForUser(ctx, userID, time.Now().Add(-24*time.Hour), time.Now().Add(24*time.Hour))
	require.NoError(t, err)
	require.Len(t, completedTodos, 1)
	assertPopulatedTodoFixture(t, &completedTodos[0], completedFixture)

	overdueTodos, err := todoRepo.GetOverdueTodosForUser(ctx, userID)
	require.NoError(t, err)
	require.Len(t, overdueTodos, 1)
	assertPopulatedTodoFixture(t, &overdueTodos[0], overdueFixture)
}

func createTestTodo(t *testing.T, ctx context.Context, repo *repository.TodoRepository, userID string) *todo.Todo {
	t.Helper()

	dueDate := time.Now().Add(24 * time.Hour)
	payload := &todo.CreateTodoPayload{
		Title:       "Test Todo",
		Description: testing_pkg.Ptr("Test todo description"),
		Priority:    testing_pkg.Ptr(todo.PriorityHigh),
		DueDate:     &dueDate,
	}

	result, err := repo.CreateTodo(ctx, userID, payload)
	require.NoError(t, err)

	return result
}

func createTestTodoWithDueDate(t *testing.T, ctx context.Context, repo *repository.TodoRepository, userID, title string, dueDate time.Time) *todo.Todo {
	t.Helper()

	payload := &todo.CreateTodoPayload{
		Title:       title,
		Description: testing_pkg.Ptr(fmt.Sprintf("%s description", title)),
		Priority:    testing_pkg.Ptr(todo.PriorityHigh),
		DueDate:     &dueDate,
	}

	result, err := repo.CreateTodo(ctx, userID, payload)
	require.NoError(t, err)

	return result
}

func createTestTodos(t *testing.T, ctx context.Context, repo *repository.TodoRepository, userID string, count int) []*todo.Todo {
	t.Helper()

	todos := make([]*todo.Todo, 0, count)

	for i := 0; i < count; i++ {
		dueDate := time.Now().Add(time.Duration(i+1) * 24 * time.Hour)
		payload := &todo.CreateTodoPayload{
			Title:       fmt.Sprintf("Test Todo %d", i+1),
			Description: testing_pkg.Ptr(fmt.Sprintf("Test todo description %d", i+1)),
			Priority:    testing_pkg.Ptr(todo.PriorityHigh),
			DueDate:     &dueDate,
		}

		result, err := repo.CreateTodo(ctx, userID, payload)
		require.NoError(t, err)
		todos = append(todos, result)

		// Add a small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	return todos
}

func setTodoStatus(t *testing.T, ctx context.Context, repo *repository.TodoRepository, userID string, todoID uuid.UUID, status todo.Status) *todo.Todo {
	t.Helper()

	payload := &todo.UpdateTodoPayload{
		ID:     todoID,
		Status: &status,
	}

	result, err := repo.UpdateTodo(ctx, userID, payload)
	require.NoError(t, err)

	return result
}

type populatedTodoFixture struct {
	Parent      *todo.Todo
	Child       *todo.Todo
	Comments    []*commentModel.Comment
	Attachments []*attachmentModel.Attachment
}

func createPopulatedTodoFixture(t *testing.T, ctx context.Context, todoRepo *repository.TodoRepository, commentRepo *repository.CommentRepository, userID string, title string, dueDate time.Time) populatedTodoFixture {
	t.Helper()

	parentPayload := &todo.CreateTodoPayload{
		Title:       title,
		Description: testing_pkg.Ptr(title + " parent description"),
		Priority:    testing_pkg.Ptr(todo.PriorityHigh),
		DueDate:     &dueDate,
	}

	parent, err := todoRepo.CreateTodo(ctx, userID, parentPayload)
	require.NoError(t, err)

	childPayload := &todo.CreateTodoPayload{
		Title:        title + " child",
		Description:  testing_pkg.Ptr(title + " child description"),
		Priority:     testing_pkg.Ptr(todo.PriorityMedium),
		ParentTodoID: &parent.ID,
	}

	child, err := todoRepo.CreateTodo(ctx, userID, childPayload)
	require.NoError(t, err)

	commentOne, err := commentRepo.CreateComment(ctx, userID, parent.ID, &commentModel.CreateCommentPayload{
		TodoID:  parent.ID,
		Content: title + " comment 1",
	})
	require.NoError(t, err)

	commentTwo, err := commentRepo.CreateComment(ctx, userID, parent.ID, &commentModel.CreateCommentPayload{
		TodoID:  parent.ID,
		Content: title + " comment 2",
	})
	require.NoError(t, err)

	attachmentOne, err := todoRepo.AddAttachment(ctx, parent.ID, userID, title+"-key-1", title+"-attachment-1.txt", 128, "text/plain")
	require.NoError(t, err)

	attachmentTwo, err := todoRepo.AddAttachment(ctx, parent.ID, userID, title+"-key-2", title+"-attachment-2.txt", 256, "text/plain")
	require.NoError(t, err)

	return populatedTodoFixture{
		Parent:      parent,
		Child:       child,
		Comments:    []*commentModel.Comment{commentOne, commentTwo},
		Attachments: []*attachmentModel.Attachment{attachmentOne, attachmentTwo},
	}
}

func assertPopulatedTodoFixture(t *testing.T, result *todo.PopulatedTodo, fixture populatedTodoFixture) {
	t.Helper()

	require.NotNil(t, result)
	assert.Equal(t, fixture.Parent.ID, result.ID)

	require.Len(t, result.Children, 1)
	assert.Equal(t, fixture.Child.ID, result.Children[0].ID)

	require.Len(t, result.Comments, len(fixture.Comments))
	assert.ElementsMatch(t, commentIDs(fixture.Comments), populatedCommentIDs(result))

	require.Len(t, result.Attachments, len(fixture.Attachments))
	assert.ElementsMatch(t, attachmentIDs(fixture.Attachments), populatedAttachmentIDs(result))
}

func commentIDs(comments []*commentModel.Comment) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(comments))
	for _, item := range comments {
		ids = append(ids, item.ID)
	}

	return ids
}

func populatedCommentIDs(todoItem *todo.PopulatedTodo) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(todoItem.Comments))
	for _, item := range todoItem.Comments {
		ids = append(ids, item.ID)
	}

	return ids
}

func attachmentIDs(attachments []*attachmentModel.Attachment) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(attachments))
	for _, item := range attachments {
		ids = append(ids, item.ID)
	}

	return ids
}

func populatedAttachmentIDs(todoItem *todo.PopulatedTodo) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(todoItem.Attachments))
	for _, item := range todoItem.Attachments {
		ids = append(ids, item.ID)
	}

	return ids
}
