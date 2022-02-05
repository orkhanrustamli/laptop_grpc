package genarator

import (
	"testing"

	pb "github.com/orkhanrustamli/pcbook/pcbook_proto/go"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestProtobufToBinaryFile(t *testing.T) {
	t.Parallel()

	binaryFile := "tmp/laptop.bin"

	laptopOne := NewLaptop()
	err := WriteProtobufToBinaryFile(laptopOne, binaryFile)
	require.NoError(t, err)

	laptopTwo := &pb.Laptop{}
	err = ReadProtobufFromBinaryFile(binaryFile, laptopTwo)
	require.NoError(t, err)
	require.True(t, proto.Equal(laptopOne, laptopTwo))
}

func TestProtobufToJSONFile(t *testing.T) {
	t.Parallel()

	jsonFile := "tmp/laptop.json"

	laptopOne := NewLaptop()
	err := WriteProtobufToJSONFile(laptopOne, jsonFile)
	require.NoError(t, err)

	laptopTwo := &pb.Laptop{}
	err = ReadProtobufFromJSONFile(jsonFile, laptopTwo)
	require.NoError(t, err)

	require.True(t, proto.Equal(laptopOne, laptopTwo))
}
