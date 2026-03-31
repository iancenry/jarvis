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
	todoRepo := repository.NewTodoRepository(testServer)
	commentRepo := repository.NewCommentRepository(testServer)
	userID := uuid.New().String()
	testTodo := createTestTodo(t, ctx, todoRepo, userID)

	t.Run("get missing comment returns not found", func(t *testing.T) {
		result, err := commentRepo.GetCommentByID(ctx, userID, uuid.New())
		require.Nil(t, result)
		assertRepositoryNotFoundError(t, err, "COMMENT_NOT_FOUND", "comment not found")
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
		err := commentRepo.DeleteComment(ctx, userID, uuid.New())
		assertRepositoryNotFoundError(t, err, "COMMENT_NOT_FOUND", "comment not found")
	})
}
