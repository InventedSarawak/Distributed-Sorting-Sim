package algorithms

import (
	"sync"
	"testing"

	"github.com/InventedSarawak/Distributed-Sorting-Sim/internal/simulator"
	"github.com/InventedSarawak/Distributed-Sorting-Sim/pkg/types"
)

func TestOddEvenSorting(t *testing.T) {
	const numNodes = 6
	// Initialize the engine for coordination
	engine := simulator.NewEngine[OddEvenPayload](numNodes)
	var wg sync.WaitGroup

	// Slice to capture final results for validation
	finalValues := make([]int, numNodes)

	for i := range numNodes {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Create a node with a reversed initial value: [5, 4, 3, 2, 1, 0]
			node := &types.Node[OddEvenPayload]{
				ID:         id,
				TotalNode:  numNodes,
				Value:      OddEvenPayload{Value: numNodes - 1 - id},
				LeftInbox:  make(chan types.Message[OddEvenPayload], 100),
				RightInbox: make(chan types.Message[OddEvenPayload], 100),
			}

			// Assign Position for boundary checks
			switch id {
			case 0:
				node.Position = types.Head
			case numNodes - 1:
				node.Position = types.Tail
			default:
				node.Position = types.Middle
			}

			// Setup synchronization buffers
			leftBuf := simulator.NewRoundBuffer(node.LeftInbox)
			rightBuf := simulator.NewRoundBuffer(node.RightInbox)

			// Note: In a real test, you'd need the transport layer active.
			// This test assumes internal logic flow.
			RunOddEven(node, engine, leftBuf, rightBuf)

			finalValues[id] = node.Value.Value
		}(i)
	}

	wg.Wait()

	// Validate: Final array should be [0, 1, 2, 3, 4, 5]
	for i := 0; i < numNodes-1; i++ {
		if finalValues[i] > finalValues[i+1] {
			t.Errorf("Sort failed! Value at index %d (%d) is greater than index %d (%d)",
				i, finalValues[i], i+1, finalValues[i+1])
		}
	}
}
