package simulator

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	"github.com/InventedSarawak/Distributed-Sorting-Sim/internal/transport"
	"github.com/InventedSarawak/Distributed-Sorting-Sim/pkg/shared"
	"github.com/InventedSarawak/Distributed-Sorting-Sim/pkg/types"
)

// SimulatorEngine manages global state and termination for a simulation of type T.
// Fix: The struct is now generic to allow methods to interact with generic Nodes.
type SimulatorEngine[T any] struct {
	TotalNodes  int
	ActiveNodes int32 // Atomic counter for termination detection
	Done        chan bool
	WaitGroup   sync.WaitGroup
}

// NewEngine initializes the engine.
func NewEngine[T any](n int) *SimulatorEngine[T] {
	return &SimulatorEngine[T]{
		TotalNodes: n,
		Done:       make(chan bool),
	}
}

// IncrementClock advances the node's logical round.
// Fix: Removed [T any] from the method signature to resolve the compiler error.
func (e *SimulatorEngine[T]) IncrementClock(n *types.Node[T]) {
	n.Round++
}

// InitialSetup orchestrates discovery and value generation.
func (e *SimulatorEngine[T]) InitialSetup(n *types.Node[T], config types.Config, leftBuf, rightBuf *RoundBuffer[T]) error {
	// Call consolidated discovery logic
	total, err := DiscoverTotalNodes(n, leftBuf, rightBuf)
	if err != nil {
		return err
	}
	e.TotalNodes = total

	// Call shared value generator
	shared.GenerateInitialValue(n, config)
	return nil
}

func SetupNode[T any](n *types.Node[T]) error {
	// 1. Start the listener in a background goroutine so neighbors can dial in.
	// The internal/transport/tcp_node.go Listen function handles the persistent dispatcher.
	go func() {
		if err := transport.Listen(n.ID, n.LeftInbox); err != nil {
			fmt.Printf("Node %d listener error: %v\n", n.ID, err)
		}
	}()

	// 2. Establish persistent links (blocking until neighbors are ready via retry logic).
	// Head and Middle nodes dial their Right neighbor.
	if n.Position != types.Tail {
		conn, err := transport.DialNeighbor(n.ID + 1)
		if err != nil {
			return fmt.Errorf("failed to link right neighbor: %w", err)
		}
		n.RightConn = conn
	}

	return nil
}

func (e *SimulatorEngine[T]) SignalStable() {
	if atomic.AddInt32(&e.ActiveNodes, -1) == 0 {
		e.Done <- true
	}
}

func (e *SimulatorEngine[T]) ResetStability() {
	atomic.StoreInt32(&e.ActiveNodes, int32(e.TotalNodes))
}

func (e *SimulatorEngine[T]) CheckTermination() bool {
	select {
	case <-e.Done:
		return true
	default:
		return false
	}
}

// DiscoverTotalNodes implements the bidirectional increment method.
// Refactor: Consolidates redundant Head/Tail and Left/Right logic.
func DiscoverTotalNodes[T any](n *types.Node[T], leftBuf, rightBuf *RoundBuffer[T]) (int, error) {
	leftDist, rightDist := -1, -1

	// Helper to send discovery seeds from boundary nodes
	sendSeed := func(conn net.Conn, dist uint64) error {
		msg := types.Message[T]{Type: types.MsgInit, SenderID: n.ID, Round: 0, Sequence: dist}
		return transport.SendMessage(conn, msg)
	}

	if n.Position == types.Head {
		leftDist = 0
		if err := sendSeed(n.RightConn, uint64(leftDist)); err != nil {
			return -1, err
		}
	}
	if n.Position == types.Tail {
		rightDist = 0
		if err := sendSeed(n.LeftConn, uint64(rightDist)); err != nil {
			return -1, err
		}
	}

	// Loop until both distances are discovered
	for leftDist == -1 || rightDist == -1 {
		if leftDist == -1 {
			msg := leftBuf.GetStepMessage(0)
			leftDist = int(msg.Sequence) + 1
			if n.Position != types.Tail {
				msg.Sequence = uint64(leftDist)
				transport.SendMessage(n.RightConn, msg)
			}
		}
		if rightDist == -1 {
			msg := rightBuf.GetStepMessage(0)
			rightDist = int(msg.Sequence) + 1
			if n.Position != types.Head {
				msg.Sequence = uint64(rightDist)
				transport.SendMessage(n.LeftConn, msg)
			}
		}
	}

	n.TotalNode = leftDist + rightDist + 1
	return n.TotalNode, nil
}
