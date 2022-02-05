package client

import (
	"context"
	"time"

	pb "github.com/orkhanrustamli/pcbook/pcbook_proto/go"
	"google.golang.org/grpc"
)

type AuthClient struct {
	service  pb.AuthServiceClient
	Username string
	Password string
}

func NewAuthClient(conn *grpc.ClientConn, username, password string) *AuthClient {
	service := pb.NewAuthServiceClient(conn)

	return &AuthClient{
		service:  service,
		Username: username,
		Password: password,
	}
}

func (authClient *AuthClient) Login() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.LoginRequest{
		Username: authClient.Username,
		Password: authClient.Password,
	}

	res, err := authClient.service.Login(ctx, req)
	if err != nil {
		return "", err
	}

	return res.AccessToken, nil
}
