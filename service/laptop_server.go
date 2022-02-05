package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/orkhanrustamli/pcbook/pcbook_proto/go"
)

const (
	maxImageSize = 1 << 20
)

type LaptopServer struct {
	Store LaptopStore
	ImageStore
	RatingStore
	pb.UnimplementedLaptopServiceServer
}

func NewLaptopServer(laptopStore LaptopStore, imageStore ImageStore, ratingStore RatingStore) *LaptopServer {
	return &LaptopServer{
		Store:       laptopStore,
		ImageStore:  imageStore,
		RatingStore: ratingStore,
	}
}

func (server *LaptopServer) CreateLaptop(
	ctx context.Context,
	req *pb.CreateLaptopRequest,
) (*pb.CreateLaptopResponse, error) {
	laptop := req.GetLaptop()
	log.Printf("Received a create-laptop request with id: %v", laptop.Id)

	if len(laptop.Id) > 0 {
		_, err := uuid.Parse(laptop.Id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Laptop ID is not a valid UUID: %v", laptop.Id)
		}
	} else {
		id, err := uuid.NewRandom()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Cannot generate a new laptop ID: %v", err)
		}
		laptop.Id = id.String()
	}

	// Emulate the context timeout and cancel
	if ctx.Err() == context.Canceled {
		return nil, logAndReturnError(status.Errorf(codes.Canceled, "Request is cancelled"))
	}

	if ctx.Err() == context.DeadlineExceeded {
		return nil, logAndReturnError(status.Errorf(codes.DeadlineExceeded, "Deadline exceeded!"))
	}

	if err := server.Store.Save(laptop); err != nil {
		code := codes.Internal
		if errors.Is(err, ErrAlreadyExists) {
			code = codes.AlreadyExists
		}

		return nil, status.Errorf(code, err.Error())
	}

	log.Printf("Laptop was saved with ID: %v", laptop.Id)

	res := &pb.CreateLaptopResponse{
		Id: laptop.Id,
	}
	return res, nil
}

func (server *LaptopServer) SearchLaptop(
	req *pb.SearchLaptopRequest,
	stream pb.LaptopService_SearchLaptopServer,
) error {
	filter := req.GetFilter()
	log.Printf("Received a search-laptop request with filter: %v", filter)

	err := server.Store.Search(
		stream.Context(),
		filter,
		func(laptop *pb.Laptop) error {
			res := &pb.SearchLaptopResponse{Laptop: laptop}

			if err := stream.Send(res); err != nil {
				return err
			}

			log.Printf("Sent laptop with id: %s", laptop.GetId())
			return nil
		},
	)
	if err != nil {
		return status.Errorf(codes.Internal, "unexpected error: %v", err)
	}

	return nil
}

func (server *LaptopServer) UploadImage(stream pb.LaptopService_UploadImageServer) error {
	req, err := stream.Recv()
	if err != nil {
		return logAndReturnError(status.Errorf(codes.Unknown, "Cannot receive image info"))
	}

	laptopId := req.GetInfo().GetLaptopId()
	imageType := req.GetInfo().GetImageType()
	log.Printf("Received an upload-image request for laptop %s with image type %s", laptopId, imageType)

	laptop, err := server.Store.Find(laptopId)
	if err != nil {
		return logAndReturnError(status.Errorf(codes.Internal, "Cannot find laptop: %v", err))
	}
	if laptop == nil {
		return logAndReturnError(status.Errorf(codes.Internal, "Laptop %s does not exists", laptopId))
	}

	imageData := bytes.Buffer{}
	imageSize := 0

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("No more chunk")
			break
		}
		if err != nil {
			return logAndReturnError(status.Errorf(codes.Unknown, "Cannot receive image chunk: %v", err))
		}

		chunk := req.GetChunkData()
		size := len(chunk)

		imageSize += size
		if imageSize > maxImageSize {
			return logAndReturnError(status.Errorf(codes.InvalidArgument, "Image size exceed maximum allowed image size (1MB)"))
		}

		if _, err := imageData.Write(chunk); err != nil {
			return status.Errorf(codes.Internal, "Cannot write chunk data: %v", err)
		}
	}

	imageId, err := server.ImageStore.Save(laptopId, imageType, imageData)
	if err != nil {
		return status.Errorf(codes.Internal, "Cannot save image to the store: %v", err)
	}

	res := &pb.UploadImageResponse{
		Id:   imageId,
		Size: uint32(imageSize),
	}

	if err := stream.SendAndClose(res); err != nil {
		return status.Errorf(codes.Unknown, "Cannot send response: %v", err)
	}

	log.Printf("Saved image with id: %s, id: %d", imageId, imageSize)
	return nil
}

func (server *LaptopServer) RateLaptop(stream pb.LaptopService_RateLaptopServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("No more data")
			break
		}
		if err != nil {
			return logAndReturnError(status.Errorf(codes.Unknown, "Cannot receive rating: %v", err))
		}

		laptopId := req.GetLaptopId()
		score := req.GetScore()

		log.Printf("Received rate-laptop request with id: %s, score: %v", laptopId, score)

		laptop, err := server.Store.Find(req.LaptopId)
		if err != nil {
			return logAndReturnError(status.Errorf(codes.Internal, "there is no laptop with id:%s in the store %v", laptopId, err))
		}
		if laptop == nil {
			return logAndReturnError(status.Errorf(codes.NotFound, "there is no laptop with id:%s in the store", laptopId))
		}

		rating := server.RatingStore.Rate(laptopId, score)
		res := &pb.RateLaptopResponse{
			LaptopId:    laptopId,
			RatingCount: uint32(rating.count),
			AvarageScre: rating.sum / float64(rating.count),
		}

		if err = stream.Send(res); err != nil {
			return logAndReturnError(status.Errorf(codes.Unknown, "cannot send response: %v", err))

		}
	}

	return nil
}

// UTILS
func logAndReturnError(err error) error {
	log.Print(err)
	return err
}
