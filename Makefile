.PHONY: proto clean build run-server run-client

# Generate protobuf files
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/printer.proto

# Clean generated files
clean:
	rm -f proto/*.pb.go

# Build binaries
build: proto
	go build -o bin/server server/main.go
	go build -o bin/client client/main.go

# Run server
run-server: proto
	go run server/main.go

# Run client
run-client: proto
	go run client/main.go
