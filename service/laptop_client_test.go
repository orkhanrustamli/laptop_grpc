package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/orkhanrustamli/pcbook/genarator"
	pb "github.com/orkhanrustamli/pcbook/pcbook_proto/go"
	"github.com/stretchr/testify/require"

	"google.golang.org/grpc"
)

func TestClientCreatelaptop(t *testing.T) {
	t.Parallel()

	store := NewInMemoryLaptopStore()

	address := startTestLaptopServer(t, store, nil, nil)
	laptopClient := startTestLaptopClient(t, address)

	laptop := genarator.NewLaptop()
	laptopId := laptop.Id
	laptopRequest := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	res, err := laptopClient.CreateLaptop(context.Background(), laptopRequest)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotEmpty(t, res.Id)
	require.Equal(t, res.Id, laptopId)

	other, err := store.Find(laptopId)
	require.NoError(t, err)
	require.NotNil(t, other)

	requireSameLaptop(t, laptop, other)

}

func TestClientSearchLaptop(t *testing.T) {
	t.Parallel()

	filter := &pb.Filter{
		MaxPriceUsd: 2000,
		MinCpuCores: 4,
		MinCpuGhz:   2.2,
		MinRam:      &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE},
	}

	store := NewInMemoryLaptopStore()
	expectedIDs := make(map[string]bool)

	for i := 0; i < 6; i++ {
		newLaptop := genarator.NewLaptop()
		switch i {
		case 0:
			newLaptop.PriceUsd = 3000
		case 1:
			newLaptop.Cpu.NumberCores = 3
		case 3:
			newLaptop.Cpu.MinGhz = 2.0
		case 4:
			newLaptop.Ram = &pb.Memory{Value: 7, Unit: pb.Memory_GIGABYTE}
		case 5:
			newLaptop.PriceUsd = 1900
			newLaptop.Cpu.NumberCores = 5
			newLaptop.Cpu.MinGhz = 2.5
			newLaptop.Ram = &pb.Memory{Value: 9, Unit: pb.Memory_GIGABYTE}
			expectedIDs[newLaptop.GetId()] = true
		}

		err := store.Save(newLaptop)
		require.NoError(t, err)
	}

	address := startTestLaptopServer(t, store, nil, nil)
	client := startTestLaptopClient(t, address)

	req := &pb.SearchLaptopRequest{Filter: filter}
	stream, err := client.SearchLaptop(context.Background(), req)
	require.NoError(t, err)

	found := 0
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}

		require.NoError(t, err)
		require.Contains(t, expectedIDs, res.Laptop.GetId())

		found++
	}

	require.Equal(t, len(expectedIDs), found)
}

func TestClientUploadImage(t *testing.T) {
	t.Parallel()

	imageFolder := "../tmp"
	imagePath := fmt.Sprintf("%s/laptop.jpg", imageFolder)
	imageType := filepath.Ext(imagePath)

	laptopStore := NewInMemoryLaptopStore()
	imageStore := NewDiskImageStore(imageFolder)

	laptop := genarator.NewLaptop()
	laptopStore.Save(laptop)

	serverAddress := startTestLaptopServer(t, laptopStore, imageStore, nil)
	client := startTestLaptopClient(t, serverAddress)

	image, err := os.Open(imagePath)
	require.NoError(t, err)
	defer image.Close()

	stream, err := client.UploadImage(context.Background())
	require.NoError(t, err)

	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId:  laptop.GetId(),
				ImageType: imageType,
			},
		},
	}

	err = stream.Send(req)
	require.NoError(t, err)

	reader := bufio.NewReader(image)
	buffer := make([]byte, 1024)
	size := 0

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		require.NoError(t, err)

		size += n
	}

	res, err := stream.CloseAndRecv()
	require.NoError(t, err)
	require.NotZero(t, res.GetId())
	require.EqualValues(t, res.GetSize(), size)

	savedImagePath := fmt.Sprintf("%s/%s%s", imageFolder, res.GetId(), imageType)
	require.FileExists(t, savedImagePath)
	require.NoError(t, os.Remove(savedImagePath))
}

func TestClientRateLaptop(t *testing.T) {
	t.Parallel()

	laptopStore := NewInMemoryLaptopStore()
	rateStore := NewInMemoryRatingStore()

	laptop := genarator.NewLaptop()
	err := laptopStore.Save(laptop)
	require.NoError(t, err)

	serverAddress := startTestLaptopServer(t, laptopStore, nil, rateStore)
	client := startTestLaptopClient(t, serverAddress)

	stream, err := client.RateLaptop(context.Background())
	require.NoError(t, err)

	scores := []float64{8, 7.5, 10}
	averages := []float64{8, 7.75, 8.5}

	n := len(scores)
	for i := 0; i < n; i++ {
		req := &pb.RateLaptopRequest{
			LaptopId: laptop.GetId(),
			Score:    scores[i],
		}

		err := stream.Send(req)
		require.NoError(t, err)
	}

	err = stream.CloseSend()
	require.NoError(t, err)

	for idx := 0; ; idx++ {
		res, err := stream.Recv()
		if err == io.EOF {
			require.Equal(t, n, idx)
			return
		}

		require.NoError(t, err)
		require.Equal(t, laptop.GetId(), res.GetLaptopId())
		require.Equal(t, uint32(idx+1), res.GetRatingCount())
		require.Equal(t, averages[idx], res.GetAvarageScre())
	}
}

// UTILITES
func startTestLaptopServer(t *testing.T, store LaptopStore, imageStore ImageStore, ratingStore RatingStore) string {
	laptopServer := NewLaptopServer(store, imageStore, ratingStore)

	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	go grpcServer.Serve(listener)

	return listener.Addr().String()
}

func startTestLaptopClient(t *testing.T, address string) pb.LaptopServiceClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	require.NoError(t, err)

	return pb.NewLaptopServiceClient(conn)
}

func requireSameLaptop(t *testing.T, laptop1 *pb.Laptop, laptop2 *pb.Laptop) {
	json1, err := genarator.ProtobufToJson(laptop1)
	require.NoError(t, err)

	json2, err := genarator.ProtobufToJson(laptop2)
	require.NoError(t, err)

	require.Equal(t, json1, json2)
}
