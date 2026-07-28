package grpc

import (
	"context"

	"github.com/MiRRoRise/auth-service/internal/usecase"
	pb "github.com/MiRRoRise/auth-service/proto/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedAuthServiceServer
	authUseCase usecase.UserUseCase
}

func NewServer(authUseCase usecase.UserUseCase) *Server {
	return &Server{
		authUseCase: authUseCase,
	}
}

func (s *Server) GetUserByID(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := s.authUseCase.GetUserByID(ctx, req.UserId)
	if err != nil || user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &pb.GetUserResponse{
		UserId:   user.ID,
		Email:    user.Email,
		IsActive: true,
	}, nil
}
