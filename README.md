# gRPC_distributed_printer

## Commands

```bash
# Generate protobuf files
make proto

# Run server (terminal 1)
go run server/main.go

# Run clients manually (terminals 2, 3, 4...)
go run client/main.go -id A -port 50052 -peers localhost:50053,localhost:50054
go run client/main.go -id B -port 50053 -peers localhost:50052,localhost:50054
go run client/main.go -id C -port 50054 -peers localhost:50052,localhost:50053

# Run with auto mode (it will open the server, then 3 clients in auto mode)
./run_auto.sh
```
