Dumb Print Server (Port 50051):

Runs in a separate terminal on port 50051
Wait passively for connections
Upon receiving a message via SendToPrinter, prints and returns confirmation
Is unaware of the existence of other clients
Does NOT participate in the mutual exclusion algorithm
Does NOT know about other clients

Smart Clients (Ports 50052, 50053, 50054, etc.):

Implements the full Ricart-Agrawala algorithm
Keeps Lamport's logical clocks up to date
Generates automatic print requests
Displays local status in real time
