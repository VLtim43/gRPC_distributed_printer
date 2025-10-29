package clock

import "sync"

type Lamport struct {
	time  int64
	mutex sync.Mutex
}

func New() *Lamport {
	return &Lamport{time: 0}
}

// Increment increments the clock and returns the new value
func (lc *Lamport) Increment() int64 {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()
	lc.time++
	return lc.time
}

// Update updates the clock based on a received timestamp
func (lc *Lamport) Update(receivedTime int64) int64 {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()
	if receivedTime > lc.time {
		lc.time = receivedTime
	}
	lc.time++
	return lc.time
}

// Get returns the current clock value without modifying it
func (lc *Lamport) Get() int64 {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()
	return lc.time
}
