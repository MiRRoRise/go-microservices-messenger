package clients

import (
	"context"
	"fmt"
	"time"

	pb "github.com/MiRRoRise/chat-service/proto/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthClient struct {
	client pb.AuthServiceClient
	conn   *grpc.ClientConn
}

func NewAuthClient(addr string) (*AuthClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect auth-service: %w", err)
	}

	client := pb.NewAuthServiceClient(conn)

	return &AuthClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *AuthClient) Close() error {
	return c.conn.Close()
}

func (c *AuthClient) GetUserByID(ctx context.Context, userID int64) (*pb.GetUserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &pb.GetUserRequest{
		UserId: userID,
	}
	return c.client.GetUserByID(ctx, req)
}

func (c *AuthClient) ValidateUser(ctx context.Context, userID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &pb.ValidateUserRequest{
		UserId: userID,
	}

	resp, err := c.client.ValidateUser(ctx, req)
	if err != nil {
		return false, fmt.Errorf("failed to validate user: %w", err)
	}

	return resp.Exists && resp.IsActive, nil
}