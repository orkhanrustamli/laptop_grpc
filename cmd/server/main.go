package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	pb "github.com/orkhanrustamli/pcbook/pcbook_proto/go"
	"github.com/orkhanrustamli/pcbook/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	secretKey     = "secret"
	tokenDuration = 15 * time.Minute
)

var accessManager = map[string][]string{
	"/pcbook.LaptopService/" + "CreateLaptop": {"admin"},
	"/pcbook.LaptopService/" + "UploadImage":  {"admin"},
	"/pcbook.LaptopService/" + "Ratelaptop":   {"admin", "user"},
}

func seedUsers(userStore service.UserStore) error {
	err := createUser(userStore, "admin1", "secret1", "admin")
	if err != nil {
		return err
	}
	return createUser(userStore, "user1", "secret2", "user")
}

func main() {
	port := flag.Int("port", 0, "Port used for gRPC server")
	flag.Parse()
	fmt.Printf("Starting server on port: %d", *port)

	userStore := service.NewInMemoryUserStore()
	jwtManager := service.NewJWTManager(secretKey, tokenDuration)
	authServer := service.NewAuthServer(userStore, jwtManager)

	err := seedUsers(userStore)
	if err != nil {
		log.Fatal("cannot seed users")
	}

	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore("img")
	ratingStore := service.NewInMemoryRatingStore()
	laptopServer := service.NewLaptopServer(laptopStore, imageStore, ratingStore)

	authInterceptor := service.NewAuthInterceptor(jwtManager, accessManager)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
		grpc.StreamInterceptor(authInterceptor.Stream()),
	)
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
	pb.RegisterAuthServiceServer(grpcServer, authServer)
	reflection.Register(grpcServer)

	add := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", add)
	if err != nil {
		log.Fatalf("Cannot start server: %v", err)
	}

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("Cannot start server: %v", err)
	}
}

func createUser(userStore service.UserStore, username, password, role string) error {
	user, err := service.NewUser(username, password, role)
	if err != nil {
		return err
	}

	return userStore.Save(user)
}
