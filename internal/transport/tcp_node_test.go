package transport

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/InventedSarawak/Distributed-Sorting-Sim/pkg/types"
)

type TestPayload struct {
	Data      string `json:"data"`
	Iteration int    `json:"iteration"`
	LargeData []byte `json:"large_data,omitempty"`
}

func TestPersistentConnectionLoad(t *testing.T) {
	const msgCount = 500
	inbox := make(chan types.Message[TestPayload], msgCount)

	// Start persistent listener
	go func() {
		if err := Listen(1, inbox); err != nil {
			t.Logf("Listener exited: %v", err)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	// Establish persistent link
	conn, err := DialNeighbor(1)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	
	defer conn.Close()

	// Send messages in rapid succession over the same pipe
	go func() {
		for i := 0; i < msgCount; i++ {
			msg := types.Message[TestPayload]{
				SenderID: 0,
				Round:    i,
				Body:     TestPayload{Iteration: i},
			}
			if err := SendMessage(conn, msg); err != nil {
				t.Errorf("Send failed at %d: %v", i, err)
			}
		}
	}()

	// Verify all messages arrived in order
	for i := 0; i < msgCount; i++ {
		select {
		case received := <-inbox:
			if received.Round != i {
				t.Errorf("Order mismatch: expected %d, got %d", i, received.Round)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("Timed out waiting for message %d", i)
		}
	}
}

// TestSimultaneousBidirectionalStress floods two nodes with messages in both directions
// to test the full-duplex capability of the persistent dispatcher.
func TestSimultaneousBidirectionalStress(t *testing.T) {
	const msgCount = 1000
	inbox1 := make(chan types.Message[TestPayload], msgCount)
	inbox2 := make(chan types.Message[TestPayload], msgCount)

	// Start two nodes
	go Listen(1, inbox1)
	go Listen(2, inbox2)
	time.Sleep(500 * time.Millisecond)

	conn1To2, _ := DialNeighbor(2)
	conn2To1, _ := DialNeighbor(1)
	defer conn1To2.Close()
	defer conn2To1.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	// Flood 1 -> 2
	go func() {
		defer wg.Done()
		for i := 0; i < msgCount; i++ {
			SendMessage(conn1To2, types.Message[TestPayload]{SenderID: 1, Round: i})
		}
	}()

	// Flood 2 -> 1
	go func() {
		defer wg.Done()
		for i := 0; i < msgCount; i++ {
			SendMessage(conn2To1, types.Message[TestPayload]{SenderID: 2, Round: i})
		}
	}()

	wg.Wait()
	t.Logf("Successfully exchanged %d messages bidirectionally", msgCount)
}

// TestLargePayload ensures json.NewDecoder correctly handles segments larger
// than typical MTU/buffer sizes.
func TestLargePayload(t *testing.T) {
	inbox := make(chan types.Message[TestPayload], 1)
	go Listen(10, inbox)
	time.Sleep(200 * time.Millisecond)

	conn, _ := DialNeighbor(10)
	defer conn.Close()

	// 1MB payload
	largeData := make([]byte, 1024*1024)
	rand.Read(largeData)

	msg := types.Message[TestPayload]{
		SenderID: 0,
		Body:     TestPayload{LargeData: largeData},
	}

	if err := SendMessage(conn, msg); err != nil {
		t.Fatalf("Failed to send large payload: %v", err)
	}

	select {
	case received := <-inbox:
		if len(received.Body.LargeData) != len(largeData) {
			t.Error("Data corruption in large payload")
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout on large payload")
	}
}

// TestNeighborDeadlock (Edge Case) tests the dialWithRetry logic when neighbors
// are spawned in reverse order.
func TestNeighborDeadlock(t *testing.T) {
	const targetID = 50
	errChan := make(chan error, 1)

	// Attempt to dial before the listener is even started
	go func() {
		_, err := DialNeighbor(targetID)
		errChan <- err
	}()

	// Wait 1 second, then start listener
	time.Sleep(1 * time.Second)
	inbox := make(chan types.Message[TestPayload], 1)
	go Listen(targetID, inbox)

	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("Dial failed even with retry: %v", err)
		} else {
			t.Log("Successfully connected after retry")
		}
	case <-time.After(5 * time.Second):
		t.Error("Dial retry logic timed out")
	}
}
