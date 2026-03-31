package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/iancenry/jarvis/internal/model/comment"
	"github.com/iancenry/jarvis/internal/repository"
	testing_pkg "github.com/iancenry/jarvis/internal/testing"
	"github.com/stretchr/testify/require"
)

func TestCommentRepository_DomainErrors(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	todoRepo := repository.NewTodoRepository(testServer.DB)
	commentRepo := repository.NewCommentRepository(testServer.DB)
	userID := uuid.New().String()
	testTodo := createTestTodo(t, ctx, todoRepo, userID)

	t.Run("get missing comment returns not found", func(t *testing.T) {
		result, err := commentRepo.GetCommentByID(ctx, userID, uuid.New())
		require.Nil(t, result)
		assertRepositoryNotFoundError(t, err, "COMMENT_NOT_FOUND", "comment not found")
	})

	t.Run("create comment for missing todo returns not found", func(t *testing.T) {
		result, err := commentRepo.CreateComment(ctx, userID, uuid.New(), &comment.CreateCommentPayload{
			TodoID:  uuid.New(),
			Content: "Test comment",
		})
		require.Nil(t, result)
		assertRepositoryNotFoundError(t, err, "TODO_NOT_FOUND", "todo not found")
	})

	t.Run("create comment for todo owned by another user returns not found", func(t *testing.T) {
		otherUsersTodo := createTestTodo(t, ctx, todoRepo, uuid.New().String())

		result, err := commentRepo.CreateComment(ctx, userID, otherUsersTodo.ID, &comment.CreateCommentPayload{
			TodoID:  otherUsersTodo.ID,
			Content: "Test comment",
		})
		require.Nil(t, result)
		assertRepositoryNotFoundError(t, err, "TODO_NOT_FOUND", "todo not found")
	})

	t.Run("get comment for wrong user returns not found", func(t *testing.T) {
		commentItem, err := commentRepo.CreateComment(ctx, userID, testTodo.ID, &comment.CreateCommentPayload{
			TodoID:  testTodo.ID,
			Content: "Test comment",
		})
		require.NoError(t, err)

		result, err := commentRepo.GetCommentByID(ctx, uuid.New().String(), commentItem.ID)
		require.Nil(t, result)
		assertRepositoryNotFoundError(t, err, "COMMENT_NOT_FOUND", "comment not found")
	})

	t.Run("update missing comment returns not found", func(t *testing.T) {
		payload := &comment.UpdateCommentPayload{
			ID:      uuid.New(),
			Content: "Updated comment",
		}

		result, err := commentRepo.UpdateComment(ctx, userID, payload)
		require.Nil(t, result)
		assertRepositoryNotFoundError(t, err, "COMMENT_NOT_FOUND", "comment not found")
	})

	t.Run("delete missing comment returns not found", func(t *testing.T) {
		deletedComment, err := commentRepo.DeleteComment(ctx, userID, uuid.New())
		require.Nil(t, deletedComment)
		assertRepositoryNotFoundError(t, err, "COMMENT_NOT_FOUND", "comment not found")
	})
}

func TestCommentRepository_DeleteComment(t *testing.T) {
	_, testServer, cleanup := testing_pkg.SetupTest(t)
	defer cleanup()

	ctx := context.Background()
	todoRepo := repository.NewTodoRepository(testServer.DB)
	commentRepo := repository.NewCommentRepository(testServer.DB)
	userID := uuid.New().String()
	testTodo := createTestTodo(t, ctx, todoRepo, userID)

	commentItem, err := commentRepo.CreateComment(ctx, userID, testTodo.ID, &comment.CreateCommentPayload{
		TodoID:  testTodo.ID,
		Content: "Test comment",
	})
	require.NoError(t, err)

	deletedComment, err := commentRepo.DeleteComment(ctx, userID, commentItem.ID)
	require.NoError(t, err)
	require.NotNil(t, deletedComment)
	require.Equal(t, commentItem.ID, deletedComment.ID)
	require.Equal(t, commentItem.Content, deletedComment.Content)

	result, err := commentRepo.GetCommentByID(ctx, userID, commentItem.ID)
	require.Nil(t, result)
	assertRepositoryNotFoundError(t, err, "COMMENT_NOT_FOUND", "comment not found")
}
