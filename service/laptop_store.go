package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/jinzhu/copier"

	pb "github.com/orkhanrustamli/pcbook/pcbook_proto/go"
)

// Errors
var (
	ErrAlreadyExists = errors.New("Already Exists")
)

type LaptopStore interface {
	Save(*pb.Laptop) error
	Find(id string) (*pb.Laptop, error)
	Search(ctx context.Context, filter *pb.Filter, found func(laptop *pb.Laptop) error) error
}

type InMemoryLaptopStore struct {
	mutex sync.RWMutex
	data  map[string]*pb.Laptop
}

func NewInMemoryLaptopStore() *InMemoryLaptopStore {
	return &InMemoryLaptopStore{
		data: make(map[string]*pb.Laptop),
	}
}

func (store *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	if store.data[laptop.Id] != nil {
		return ErrAlreadyExists
	}

	tmp, err := deepCopy(laptop)
	if err != nil {
		return err
	}

	store.data[tmp.Id] = tmp
	return nil
}

func (store *InMemoryLaptopStore) Find(id string) (*pb.Laptop, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	laptop := store.data[id]
	if laptop == nil {
		return nil, nil
	}

	return deepCopy(laptop)

}

func (store *InMemoryLaptopStore) Search(
	ctx context.Context,
	filter *pb.Filter,
	found func(laptop *pb.Laptop,
	) error) error {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	for _, laptop := range store.data {
		if ctx.Err() == context.Canceled || ctx.Err() == context.DeadlineExceeded {
			log.Println("Context is cancelled or timed out!")
			return nil
		}

		if isQualified(filter, laptop) {
			other, err := deepCopy(laptop)
			if err != nil {
				return err
			}

			err = found(other)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func isQualified(filter *pb.Filter, laptop *pb.Laptop) bool {
	if laptop.PriceUsd > filter.MaxPriceUsd {
		return false
	}

	if laptop.Cpu.NumberCores < filter.MinCpuCores {
		return false
	}

	if laptop.Cpu.MinGhz < filter.MinCpuGhz {
		return false
	}

	if toBit(laptop.Ram) < toBit(filter.MinRam) {
		return false
	}

	return true
}

func toBit(memory *pb.Memory) uint64 {
	value := memory.Value

	switch memory.Unit {
	case pb.Memory_BIT:
		return value
	case pb.Memory_BYTE:
		return value << 3
	case pb.Memory_KILOBYTE:
		return value << 13
	case pb.Memory_MEGABYTE:
		return value << 23
	case pb.Memory_GIGABYTE:
		return value << 33
	case pb.Memory_TERABYTE:
		return value << 43
	default:
		return 0
	}
}

func deepCopy(laptop *pb.Laptop) (*pb.Laptop, error) {
	other := &pb.Laptop{}
	if err := copier.Copy(other, laptop); err != nil {
		return nil, fmt.Errorf("cannot copy laptop data: %v", err)
	}

	return other, nil
}
