package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"distributed-printer/clock"
	pb "distributed-printer/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// State represents the client's state in the Ricart-Agrawala algorithm
type State int

const (
	RELEASED State = iota // Not interested in critical section
	WANTED                // Wants to enter critical section
	HELD                  // Currently in critical section
)

// DeferredRequest represents a request that was received while in critical section
type DeferredRequest struct {
	ClientID  string
	Timestamp int64
}

// MutexClient manages the Ricart-Agrawala mutual exclusion
type MutexClient struct {
	pb.UnimplementedClientServiceServer

	config         *Config
	lamportClock   *clock.Lamport
	printClient    pb.PrintServiceClient
	peerClients    map[string]pb.ClientServiceClient      // addr -> client
	clientIDToAddr map[string]string                       // clientID -> addr

	// Ricart-Agrawala state
	state          State
	requestTime    int64
	replyCount     int
	deferredQueue  []DeferredRequest

	mu             sync.Mutex
	replyCond      *sync.Cond
}

// NewMutexClient creates a new mutex client
func NewMutexClient(config *Config, lamportClock *clock.Lamport, printClient pb.PrintServiceClient) *MutexClient {
	mc := &MutexClient{
		config:         config,
		lamportClock:   lamportClock,
		printClient:    printClient,
		peerClients:    make(map[string]pb.ClientServiceClient),
		clientIDToAddr: make(map[string]string),
		state:          RELEASED,
		deferredQueue:  make([]DeferredRequest, 0),
	}
	mc.replyCond = sync.NewCond(&mc.mu)
	return mc
}

// StartServer starts the gRPC server to listen for peer requests
func (mc *MutexClient) StartServer() error {
	lis, err := net.Listen("tcp", ":"+mc.config.ClientPort)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterClientServiceServer(grpcServer, mc)

	log.Printf("Client gRPC server listening on port %s", mc.config.ClientPort)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	return nil
}

// RequestAccess handles incoming access requests from peers (Ricart-Agrawala)
func (mc *MutexClient) RequestAccess(ctx context.Context, req *pb.AccessRequest) (*pb.AccessReply, error) {
	// Update Lamport clock with received timestamp
	mc.lamportClock.Update(req.Timestamp)

	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Learn the mapping from client ID to connection (from gRPC peer info)
	// We'll build this mapping when we receive requests from peers
	// This allows us to send targeted replies later

	log.Printf("Received access request from %s with timestamp %d (my state: %v, my requestTime: %d)",
		req.ClientId, req.Timestamp, mc.state, mc.requestTime)

	// Ricart-Agrawala decision logic:
	// Reply immediately if:
	//   1. We're in RELEASED state (not interested), OR
	//   2. We're in WANTED state but the requester has priority:
	//      - Lower timestamp wins, OR
	//      - Same timestamp but lower client ID wins
	shouldReplyNow := mc.state == RELEASED ||
		(mc.state == WANTED && (req.Timestamp < mc.requestTime ||
			(req.Timestamp == mc.requestTime && req.ClientId < mc.config.ClientID)))

	if shouldReplyNow {
		// Reply immediately with permission
		timestamp := mc.lamportClock.Increment()
		log.Printf("Granting access immediately to %s", req.ClientId)
		return &pb.AccessReply{
			ClientId:  mc.config.ClientID,
			Timestamp: timestamp,
			Granted:   true,
		}, nil
	}

	// Defer the reply - we're in HELD or we have higher priority
	log.Printf("Deferring reply to %s (we have priority or in critical section)", req.ClientId)
	mc.deferredQueue = append(mc.deferredQueue, DeferredRequest{
		ClientID:  req.ClientId,
		Timestamp: req.Timestamp,
	})

	// Send a reply but mark as not granted (client will wait)
	timestamp := mc.lamportClock.Increment()
	return &pb.AccessReply{
		ClientId:  mc.config.ClientID,
		Timestamp: timestamp,
		Granted:   false,
	}, nil
}

// ReplyAccess handles incoming access replies from peers
func (mc *MutexClient) ReplyAccess(ctx context.Context, reply *pb.AccessReply) (*pb.Empty, error) {
	// Update Lamport clock
	mc.lamportClock.Update(reply.Timestamp)

	mc.mu.Lock()
	defer mc.mu.Unlock()

	log.Printf("Received access reply from %s (granted: %v)", reply.ClientId, reply.Granted)

	// Only count granted replies
	if reply.Granted {
		mc.replyCount++
		log.Printf("Reply count: %d/%d", mc.replyCount, len(mc.config.PeerAddresses))

		// If we've received all replies, signal that we can proceed
		if mc.replyCount == len(mc.config.PeerAddresses) {
			log.Printf("All replies received! Signaling to proceed...")
			mc.replyCond.Broadcast()
		}
	}

	return &pb.Empty{}, nil
}

// ConnectToPeers establishes connections to all peer clients
func (mc *MutexClient) ConnectToPeers() error {
	for _, peerAddr := range mc.config.PeerAddresses {
		conn, err := grpc.NewClient(peerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return fmt.Errorf("failed to connect to peer %s: %v", peerAddr, err)
		}
		mc.peerClients[peerAddr] = pb.NewClientServiceClient(conn)
		log.Printf("Connected to peer at %s", peerAddr)
	}
	return nil
}

// BroadcastRequest sends access request to all peers
func (mc *MutexClient) BroadcastRequest() {
	mc.mu.Lock()
	mc.state = WANTED
	mc.requestTime = mc.lamportClock.Increment()
	mc.replyCount = 0
	requestTime := mc.requestTime
	mc.mu.Unlock()

	log.Printf("Broadcasting access request with timestamp %d to %d peers",
		requestTime, len(mc.config.PeerAddresses))

	// Send request to all peers in parallel
	var wg sync.WaitGroup
	for peerAddr, peerClient := range mc.peerClients {
		wg.Add(1)
		go func(addr string, client pb.ClientServiceClient) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			reply, err := client.RequestAccess(ctx, &pb.AccessRequest{
				ClientId:  mc.config.ClientID,
				Timestamp: requestTime,
			})

			if err != nil {
				log.Printf("Failed to send request to %s: %v", addr, err)
				return
			}

			// Learn the mapping: this reply's ClientId comes from this address
			mc.mu.Lock()
			mc.clientIDToAddr[reply.ClientId] = addr
			mc.mu.Unlock()

			// If granted, send it back to ourselves via ReplyAccess
			if reply.Granted {
				mc.ReplyAccess(context.Background(), reply)
			}
		}(peerAddr, peerClient)
	}

	wg.Wait()
	log.Printf("Finished broadcasting request to all peers")
}

// WaitForReplies waits until all peers have granted permission
func (mc *MutexClient) WaitForReplies() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Wait until we have all replies
	for mc.replyCount < len(mc.config.PeerAddresses) {
		log.Printf("Waiting for replies... (%d/%d)", mc.replyCount, len(mc.config.PeerAddresses))
		mc.replyCond.Wait()
	}

	// Enter critical section
	mc.state = HELD
	log.Printf("Permission granted! Entering critical section")
}

// ReleaseCriticalSection releases the critical section and replies to deferred requests
func (mc *MutexClient) ReleaseCriticalSection() {
	mc.mu.Lock()

	log.Printf("Releasing critical section")
	mc.state = RELEASED

	// Send replies to all deferred requests
	deferredRequests := mc.deferredQueue
	mc.deferredQueue = make([]DeferredRequest, 0)

	mc.mu.Unlock()

	// Send replies in separate goroutines
	for _, req := range deferredRequests {
		go func(clientID string) {
			// Find the specific peer address for this client ID
			mc.mu.Lock()
			peerAddr, exists := mc.clientIDToAddr[clientID]
			mc.mu.Unlock()

			if !exists {
				log.Printf("Warning: No address found for client %s, cannot send deferred reply", clientID)
				return
			}

			// Get the peer client for this specific address
			peerClient, exists := mc.peerClients[peerAddr]
			if !exists {
				log.Printf("Warning: No connection found for address %s (client %s)", peerAddr, clientID)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			timestamp := mc.lamportClock.Increment()
			log.Printf("Sending deferred reply to %s at %s", clientID, peerAddr)

			_, err := peerClient.ReplyAccess(ctx, &pb.AccessReply{
				ClientId:  mc.config.ClientID,
				Timestamp: timestamp,
				Granted:   true,
			})

			if err != nil {
				log.Printf("Failed to send deferred reply to %s: %v", clientID, err)
			}
		}(req.ClientID)
	}

	log.Printf("Sent %d deferred replies", len(deferredRequests))
}

// PrintWithMutex performs a print operation with mutual exclusion
func (mc *MutexClient) PrintWithMutex(message string) error {
	log.Printf("\n========== REQUESTING CRITICAL SECTION ==========")

	// Step 1: Broadcast request to all peers
	mc.BroadcastRequest()

	// Step 2: Wait for all replies
	mc.WaitForReplies()

	// Step 3: Enter critical section - print to server
	log.Printf("\n========== IN CRITICAL SECTION ==========")
	timestamp := mc.lamportClock.Increment()
	fmt.Printf("\n[Lamport Clock: %d] Sending print request to server...\n", timestamp)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := mc.printClient.Print(ctx, &pb.PrintRequest{
		Message:  message,
		ClientId: mc.config.ClientID,
	})
	cancel()

	if err != nil {
		log.Printf("Failed to print: %v", err)
		mc.ReleaseCriticalSection()
		return err
	}

	fmt.Printf("[Lamport Clock: %d] Server: %s\n", mc.lamportClock.Get(), resp.Result)

	// Step 4: Release critical section and send deferred replies
	log.Printf("\n========== RELEASING CRITICAL SECTION ==========")
	mc.ReleaseCriticalSection()

	return nil
}

// StartAutomaticRequests starts a background goroutine that generates print requests automatically
func (mc *MutexClient) StartAutomaticRequests(ctx context.Context) {
	ticker := time.NewTicker(mc.config.RequestInterval)
	defer ticker.Stop()

	requestCounter := 1

	log.Printf("Starting automatic print request generation (interval: %v)", mc.config.RequestInterval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping automatic request generation")
			return
		case <-ticker.C:
			message := fmt.Sprintf("Auto-generated print request #%d", requestCounter)
			requestCounter++

			log.Printf("\n>>> AUTOMATIC REQUEST: %s", message)

			// Perform print with mutual exclusion
			if err := mc.PrintWithMutex(message); err != nil {
				log.Printf("Automatic request failed: %v", err)
			}
		}
	}
}

func main() {
	// Parse configuration
	config, err := ParseConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	lamportClock := clock.New()

	fmt.Printf("========== CLIENT %s ==========\n", config.ClientID)
	fmt.Println()

	log.Printf("Client ID: %s", config.ClientID)
	log.Printf("Listening port: %s", config.ClientPort)
	log.Printf("Print server: %s", config.PrintServerAddr)
	log.Printf("Peers: %v", config.PeerAddresses)
	log.Printf("Request interval: %v", config.RequestInterval)
	fmt.Println()

	// Connect to print server
	conn, err := grpc.NewClient(config.PrintServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to print server: %v", err)
	}
	defer conn.Close()

	printClient := pb.NewPrintServiceClient(conn)

	// Create mutex client
	mutexClient := NewMutexClient(config, lamportClock, printClient)

	// Connect to peers
	log.Printf("Connecting to peers...")
	if err := mutexClient.ConnectToPeers(); err != nil {
		log.Fatalf("Failed to connect to peers: %v", err)
	}

	// Start gRPC server for peer requests
	if err := mutexClient.StartServer(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Give peers a moment to start their servers
	time.Sleep(2 * time.Second)

	log.Printf("Connected to print server as CLIENT %s", config.ClientID)
	log.Println("Using Ricart-Agrawala mutual exclusion algorithm")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handler for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\nShutting down client gracefully...")
		cancel()
		os.Exit(0)
	}()

	if config.AutoMode {
		// Automatic mode: generate requests at regular intervals
		log.Printf("AUTOMATIC MODE enabled (interval: %v)", config.RequestInterval)
		log.Println("Press Ctrl+C to stop")
		fmt.Println()

		mutexClient.StartAutomaticRequests(ctx)
	} else {
		// Manual mode: accept user input
		log.Println("MANUAL MODE: Type messages to print (Ctrl+C to exit)")
		fmt.Println()

		scanner := bufio.NewScanner(os.Stdin)

		for {
			fmt.Print("> ")
			if !scanner.Scan() {
				break
			}

			message := strings.TrimSpace(scanner.Text())
			if message == "" {
				continue
			}

			// Print with mutual exclusion
			if err := mutexClient.PrintWithMutex(message); err != nil {
				log.Printf("Error: %v", err)
			}

			fmt.Println()
		}
	}
}
