package genarator

import (
	"github.com/golang/protobuf/ptypes"
	pb "github.com/orkhanrustamli/pcbook/pcbook_proto/go"
)

func NewKeyboard() *pb.Keyboard {
	return &pb.Keyboard{
		Layout:  randomKeyboardLayout(),
		Backlit: randomBool(),
	}
}
func NewCPU() *pb.CPU {
	brand := randomCPUBrand()
	name := randomCPUName(brand)

	numberCores := randomIntRange(2, 8)
	numberThreads := randomIntRange(numberCores, 12)

	minGhz := randomFloat64Range(2.0, 3.5)
	maxGhz := randomFloat64Range(minGhz, 5.0)

	return &pb.CPU{
		Brand:         brand,
		Name:          name,
		NumberCores:   uint32(numberCores),
		NumberThreads: uint32(numberThreads),
		MinGhz:        minGhz,
		MaxGhz:        maxGhz,
	}
}

func NewGPU() *pb.GPU {
	brand := randomGPUBrand()
	name := randomGPUName(brand)

	minGhz := randomFloat64Range(1.0, 1.5)
	maxGhz := randomFloat64Range(minGhz, 2.0)

	memGB := randomIntRange(2, 6)

	return &pb.GPU{
		Brand:  brand,
		Name:   name,
		MinGhz: minGhz,
		MaxGhz: maxGhz,
		Memory: &pb.Memory{
			Value: uint64(memGB),
			Unit:  pb.Memory_GIGABYTE,
		},
	}
}

func NewRam() *pb.Memory {
	memGB := randomIntRange(4, 64)

	return &pb.Memory{
		Value: uint64(memGB),
		Unit:  pb.Memory_GIGABYTE,
	}
}

func NewSSD() *pb.Storage {
	memGB := randomIntRange(128, 1024)

	return &pb.Storage{
		Driver: pb.Storage_SSD,
		Memory: &pb.Memory{
			Value: uint64(memGB),
			Unit:  pb.Memory_GIGABYTE,
		},
	}
}

func NewHDD() *pb.Storage {
	memGB := randomIntRange(1, 6)

	return &pb.Storage{
		Driver: pb.Storage_HDD,
		Memory: &pb.Memory{
			Value: uint64(memGB),
			Unit:  pb.Memory_TERABYTE,
		},
	}
}

func newScreen() *pb.Screen {
	return &pb.Screen{
		SizeInch:    randomFloat32Range(13, 17),
		Resolution:  randomScreenResolution(),
		Panel:       randomScreenPanel(),
		Touchscreen: randomBool(),
	}
}

func NewLaptop() *pb.Laptop {
	brand := randomLaptopBrand()
	name := randomLaptopName(brand)

	return &pb.Laptop{
		Id:       randomID(),
		Brand:    brand,
		Name:     name,
		Cpu:      NewCPU(),
		Ram:      NewRam(),
		Gpu:      []*pb.GPU{NewGPU()},
		Storages: []*pb.Storage{NewSSD(), NewHDD()},
		Screen:   newScreen(),
		Keyboard: NewKeyboard(),
		Weight: &pb.Laptop_WeightKg{
			WeightKg: randomFloat64Range(1.0, 3.0),
		},
		PriceUsd:    randomFloat64Range(1500, 3500),
		ReleaseYear: uint32(randomIntRange(2015, 2021)),
		UpdatedAt:   ptypes.TimestampNow(),
	}
}
