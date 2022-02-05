package client

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type AuthInterceptor struct {
	authClient     *AuthClient
	allowedMethods map[string]bool
	accessToken    string
}

func NewAuthInterceptor(authClient *AuthClient, allowedMetgods map[string]bool, refreshDuration time.Duration) (*AuthInterceptor, error) {
	authInterceptor := &AuthInterceptor{
		authClient:     authClient,
		allowedMethods: allowedMetgods,
	}

	err := authInterceptor.scheduleRefreshToken(refreshDuration)
	if err != nil {
		return nil, err
	}

	return authInterceptor, nil
}

func (authInterceptor *AuthInterceptor) Unary(
	ctx context.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	log.Printf("---> Unary Clien Interceptor\n")

	if authInterceptor.allowedMethods[method] {
		return invoker(authInterceptor.attachToken(ctx), method, req, reply, cc, opts...)
	}

	return invoker(ctx, method, req, reply, cc, opts...)
}

func (authInterceptor *AuthInterceptor) Stream(
	ctx context.Context,
	desc *grpc.StreamDesc,
	cc *grpc.ClientConn,
	method string,
	streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	log.Printf("---> Stream Clien Interceptor\n")

	if authInterceptor.allowedMethods[method] {
		return streamer(authInterceptor.attachToken(ctx), desc, cc, method, opts...)
	}

	return streamer(ctx, desc, cc, method, opts...)
}

func (authInterceptor *AuthInterceptor) attachToken(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", authInterceptor.accessToken)
}

func (authInterceptor *AuthInterceptor) scheduleRefreshToken(refreshDuration time.Duration) error {
	err := authInterceptor.refreshToken()
	if err != nil {
		return nil
	}

	go func() {
		waitTime := refreshDuration
		for {
			time.Sleep(waitTime)
			err := authInterceptor.refreshToken()
			if err != nil {
				waitTime = time.Second
			} else {
				waitTime = refreshDuration
			}
		}
	}()

	return nil
}

func (authInterceptor *AuthInterceptor) refreshToken() error {
	accessToken, err := authInterceptor.authClient.Login()
	if err != nil {
		return err
	}

	log.Println("Access token refreshed")
	authInterceptor.accessToken = accessToken
	return nil
}
