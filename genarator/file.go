package genarator

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

func WriteProtobufToBinaryFile(message proto.Message, filename string) error {
	data, err := proto.Marshal(message)
	if err != nil {
		return fmt.Errorf("Cannot marshal proto message to binary: %v", err)
	}

	err = ioutil.WriteFile(filename, data, 0666)
	if err != nil {
		return fmt.Errorf("Cannot write binary data to file: %v", err)

	}

	return nil
}

func ReadProtobufFromBinaryFile(filename string, message proto.Message) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Cannot read from binary file: %v", err)
	}

	if err := proto.Unmarshal(data, message); err != nil {
		return fmt.Errorf("Cannot unmarshal binary data to proto message: %v", err)
	}

	return nil
}

func WriteProtobufToJSONFile(message proto.Message, filename string) error {
	dataString, err := ProtobufToJson(message)
	if err != nil {
		return fmt.Errorf("Cannot marshal proto message to string: %v", err)
	}

	if err := ioutil.WriteFile(filename, []byte(dataString), 0644); err != nil {
		return fmt.Errorf("Cannot write string data to json file: %v", err)
	}

	return nil
}

func ReadProtobufFromJSONFile(filename string, message proto.Message) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Cannot read from binary file: %v", err)
	}

	if err := jsonpb.Unmarshal(strings.NewReader(string(data)), message); err != nil {
		return fmt.Errorf("Cannot unmarshal binary data to proto message: %v", err)
	}

	return nil
}

// Utils
func ProtobufToJson(message proto.Message) (string, error) {
	marshaler := jsonpb.Marshaler{
		EnumsAsInts:  false,
		EmitDefaults: true,
		Indent:       "  ",
		OrigName:     true,
	}

	return marshaler.MarshalToString(message)
}
