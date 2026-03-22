package repository

import "github.com/iancenry/jarvis/internal/server"

type CommentRepository struct {
	server *server.Server
}

func NewCommentRepository(s *server.Server) *CommentRepository {
	return &CommentRepository{
		server: s,
	}
}

