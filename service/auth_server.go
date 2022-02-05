package service

import (
	"context"
	"log"

	pb "github.com/orkhanrustamli/pcbook/pcbook_proto/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	userStore  UserStore
	jwtManager *JWTManager
	pb.UnimplementedAuthServiceServer
}

func NewAuthServer(userStore UserStore, jwtManager *JWTManager) *AuthServer {
	return &AuthServer{
		userStore:  userStore,
		jwtManager: jwtManager,
	}
}

func (authServer *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	username := req.GetUsername()
	password := req.GetPassword()
	log.Printf("Received a login request with username:%s", username)

	user := authServer.userStore.Find(username)
	if user == nil {
		return nil, status.Errorf(codes.NotFound, "user with username:%s not found", username)
	}

	if !user.IsCorrectPassword(password) {
		return nil, status.Errorf(codes.InvalidArgument, "password provided for user:%s is incorrect", password)
	}

	accessToken, err := authServer.jwtManager.Generate(username, user.Role)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot generate a token for user:%s -- err: %v", username, err)
	}

	res := &pb.LoginResponse{
		AccessToken: accessToken,
	}
	return res, nil
}
