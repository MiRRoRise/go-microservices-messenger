package grpc

import (
	"context"

	"github.com/MiRRoRise/auth-service/internal/usecase"
	pb "github.com/MiRRoRise/auth-service/proto/auth"
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
	if err != nil {
		return &pb.GetUserResponse{
			UserId:    0,
			Email:     "",
			IsActive:  false,
			CreatedAt: "",
		}, nil
	}

	return &pb.GetUserResponse{
		UserId:    user.ID,
		Email:     user.Email,
		IsActive:  true,
		CreatedAt: user.CreatedAt.String(),
	}, nil
}

func (s *Server) ValidateUser(ctx context.Context, req *pb.ValidateUserRequest) (*pb.ValidateUserResponse, error) {
	user, err := s.authUseCase.GetUserByID(ctx, req.UserId)
	if err != nil {
		return &pb.ValidateUserResponse{
			Exists: false,
			IsActive: false,
		}, nil
	}

	return &pb.ValidateUserResponse{
		Exists: user != nil,
		IsActive: true,
	}, nil
}