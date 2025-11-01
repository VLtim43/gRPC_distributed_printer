# gRPC_distributed_printer

A distributed printing system using gRPC for inter-process communication, Ricart-Agrawala algorithm for distributed mutual exclusion, and Lamport logical clocks for event ordering.

The print server (port 50051) is stateless and doesn't participate in coordination. the clients run both as gRPC servers and clients here, as they receive mutual exclusion requests from peers (that you have to pass as a parameter) while simultaneously requesting access themselves.
When a client wants to print, it broadcasts a timestamped request to all (known) peers. Each peer either grants permission immediately (if not interested or has lower priority) or defers the reply (if currently printing or has higher priority based on timestamp comparison). Once all peers grant permission, the client accesses the printer. After printing, deferred replies are sent to waiting clients. Lamport clocks ensure causal ordering: clocks increment on local events and update to max(local, received) + 1 on message receipt.

You can ctrl+c on the cliens to stop them while the terminal is still open. So you can check their messages queues and stuff. I had problems with the main server being still open even after it's terminal killed, so I had to use this kill_server.sh to kill the process.

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
