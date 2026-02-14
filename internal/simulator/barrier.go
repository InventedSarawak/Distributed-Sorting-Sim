package simulator

import (
	"sync"

	"github.com/InventedSarawak/Distributed-Sorting-Sim/internal/transport"
	"github.com/InventedSarawak/Distributed-Sorting-Sim/pkg/types"
)

// RoundBuffer manages out-of-order messages for a specific neighbor connection.
type RoundBuffer[T any] struct {
	mu         sync.Mutex
	futureMsgs map[int]types.Message[T] // Stores messages for future rounds
	inbox      chan types.Message[T]    // Connection to the transport dispatcher
}

func NewRoundBuffer[T any](inbox chan types.Message[T]) *RoundBuffer[T] {
	return &RoundBuffer[T]{
		futureMsgs: make(map[int]types.Message[T]),
		inbox:      inbox,
	}
}

// GetStepMessage retrieves a message for a specific round, blocking if necessary.
func (rb *RoundBuffer[T]) GetStepMessage(targetRound int) types.Message[T] {
	for {
		rb.mu.Lock()
		// Check if the message was already buffered from a previous dispatcher push
		if msg, found := rb.futureMsgs[targetRound]; found {
			delete(rb.futureMsgs, targetRound)
			rb.mu.Unlock()
			return msg
		}
		rb.mu.Unlock()

		// If not in buffer, wait for the next message from the Inbox Dispatcher
		msg := <-rb.inbox

		if msg.Round == targetRound {
			return msg
		}

		// Store future-round messages to prevent data loss during high-load runs
		rb.mu.Lock()
		rb.futureMsgs[msg.Round] = msg
		rb.mu.Unlock()
	}
}

// WaitForNeighbors acts as a local barrier for the current round logic.
func WaitForNeighbors[T any](n *types.Node[T], currentRound int, leftBuf, rightBuf *RoundBuffer[T]) (leftMsg, rightMsg *types.Message[T]) {
	var wg sync.WaitGroup

	// Fetch from Left neighbor if not the Head node
	if n.Position != types.Head {
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg := leftBuf.GetStepMessage(currentRound)
			leftMsg = &msg
		}()
	}

	// Fetch from Right neighbor if not the Tail node
	if n.Position != types.Tail {
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg := rightBuf.GetStepMessage(currentRound)
			rightMsg = &msg
		}()
	}

	wg.Wait()
	return
}

func BroadcastTermination[T any](n *types.Node[T]) {
	termMsg := types.Message[T]{
		Type:     types.MsgTerm,
		SenderID: n.ID,
		Round:    n.Round,
	}

	if n.Position != types.Tail {
		// Send to Right
		// (Assume SendPersistent is available from transport)
		_ = transport.SendMessage(n.RightConn, termMsg)
	}
	if n.Position != types.Head {
		// Send to Left
		// (Assume SendPersistent is available from transport)
		_ = transport.SendMessage(n.LeftConn, termMsg)
	}
}
