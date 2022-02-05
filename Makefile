gen:
	protoc --proto_path=pcbook_proto pcbook_proto/*.proto --go_out=:pcbook_proto/go --go-grpc_out=:pcbook_proto/go

clean:
	rm pcbook_proto/go/*

server:
	go run cmd/server/main.go -port 8080

client:
	go run cmd/client/main.go -address 0.0.0.0:8080