package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/orkhanrustamli/pcbook/client"
	"github.com/orkhanrustamli/pcbook/genarator"
	pb "github.com/orkhanrustamli/pcbook/pcbook_proto/go"

	"google.golang.org/grpc"
)

const (
	username        = "admin1"
	password        = "secret1"
	refreshDuration = 30 * time.Second
)

func authMethods() map[string]bool {
	const laptopServicePath = "/pcbook.LaptopService/"

	return map[string]bool{
		laptopServicePath + "CreateLaptop": true,
		laptopServicePath + "UploadImage":  true,
		laptopServicePath + "RateLaptop":   true,
	}
}

func main() {
	serverAddress := flag.String("address", "", "gRPC server address")
	flag.Parse()
	fmt.Printf("Dialing gRPC server on address: %v", *serverAddress)

	cc1, err := grpc.Dial(*serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Cannot dial the gRPC server: %v", err)
	}

	authClient := client.NewAuthClient(cc1, username, password)
	authInterceptor, err := client.NewAuthInterceptor(authClient, authMethods(), refreshDuration)
	if err != nil {
		log.Fatalf("cannot create auth interceptor: %v", err)
	}

	cc2, err := grpc.Dial(
		*serverAddress,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(authInterceptor.Unary),
		grpc.WithStreamInterceptor(authInterceptor.Stream),
	)
	if err != nil {
		log.Fatalf("Cannot dial the gRPC server: %v", err)

	}

	laptopClient := client.NewLaptopClient(cc2)
	testRateLaptop(laptopClient)

	// testImageUpload(laptopClient)
	// testSearchLaptop(laptopClient)

}

func testCreateLaptop(laptopClient *client.LaptopClient) {
	laptopClient.CreateLaptop(genarator.NewLaptop())
}

func testSearchLaptop(laptopClient *client.LaptopClient) {
	for i := 0; i < 10; i++ {
		laptopClient.CreateLaptop(genarator.NewLaptop())
	}

	filter := &pb.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 4,
		MinCpuGhz:   2.5,
		MinRam:      &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE},
	}
	laptopClient.SearchLaptop(filter)
}

func testImageUpload(laptopClient *client.LaptopClient) {
	laptop := genarator.NewLaptop()
	laptopClient.CreateLaptop(laptop)
	laptopClient.UploadImage(laptop.GetId(), "tmp/laptop.jpg")
}

func testRateLaptop(laptopClient *client.LaptopClient) {
	n := 3
	laptopIds := make([]string, n)
	scores := make([]float64, n)

	for i := 0; i < n; i++ {
		laptop := genarator.NewLaptop()
		laptopIds[i] = laptop.GetId()
		laptopClient.CreateLaptop(laptop)
	}

	for {
		fmt.Print("rate laptop (y/n)? ")
		var answer string
		fmt.Scan(&answer)

		if strings.ToLower(answer) != "y" {
			break
		}

		for i := 0; i < n; i++ {
			scores[i] = genarator.RandomLaptopScore()
		}

		err := laptopClient.RateLaptop(laptopIds, scores)
		if err != nil {
			log.Fatal(err)
		}
	}
}
