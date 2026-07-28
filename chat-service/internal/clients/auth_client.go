package clients

import (
	"context"
	"fmt"
	"time"

	pb "github.com/MiRRoRise/chat-service/proto/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type AuthClient interface {
	GetUserByID(ctx context.Context, userID int64) (*pb.GetUserResponse, error)
	ValidateUser(ctx context.Context, userID int64) (bool, error)
	Close() error
}

type authClient struct {
	client pb.AuthServiceClient
	conn   *grpc.ClientConn
}

func NewAuthClient(addr string) (*authClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect auth-service: %w", err)
	}

	client := pb.NewAuthServiceClient(conn)

	return &authClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *authClient) Close() error {
	return c.conn.Close()
}

func (c *authClient) GetUserByID(ctx context.Context, userID int64) (*pb.GetUserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &pb.GetUserRequest{
		UserId: userID,
	}
	return c.client.GetUserByID(ctx, req)
}

func (c *authClient) ValidateUser(ctx context.Context, userID int64) (bool, error) {
	resp, err := c.GetUserByID(ctx, userID)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to validate user: %w", err)
	}

	return resp != nil && resp.IsActive, nil
}
