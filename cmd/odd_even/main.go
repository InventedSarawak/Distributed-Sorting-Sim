package main

import (
	"flag"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/InventedSarawak/Distributed-Sorting-Sim/internal/algorithms"
	"github.com/InventedSarawak/Distributed-Sorting-Sim/internal/simulator"
	"github.com/InventedSarawak/Distributed-Sorting-Sim/internal/transport"
	"github.com/InventedSarawak/Distributed-Sorting-Sim/pkg/shared"
	"github.com/InventedSarawak/Distributed-Sorting-Sim/pkg/types"
)

func main() {
	nodeCount := flag.Uint("node-count", 10, "Number of Nodes") // Default to 10 for debug
	inputTypeStr := flag.String("input-type", "random", "Input type")
	flag.Parse()

	fmt.Printf("--- Distributed Sorting Simulator ---\n")
	fmt.Printf("Nodes: %d | Type: %s\n", *nodeCount, *inputTypeStr)

	// ... (Input type switch logic same as before) ...
	var inputType types.InputType
	switch strings.ToLower(*inputTypeStr) {
	case "random":
		inputType = types.Random
	case "sorted":
		inputType = types.Sorted
	case "reverse":
		inputType = types.Reverse
	default:
		panic("Invalid input type")
	}

	cfg := types.Config{
		NodeCount: uint(*nodeCount),
		InputType: inputType,
		Algorithm: types.OddEven,
	}

	engine := simulator.NewEngine[algorithms.OddEvenPayload](int(*nodeCount))
	var wg sync.WaitGroup

	for i := 0; i < int(*nodeCount); i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			node := &types.Node[algorithms.OddEvenPayload]{
				ID:        id,
				TotalNode: int(*nodeCount),
				// Buffer sizes increased to prevent deadlocks
				LeftInbox:  make(chan types.Message[algorithms.OddEvenPayload], 500),
				RightInbox: make(chan types.Message[algorithms.OddEvenPayload], 500),
			}

			if id == 0 {
				node.Position = types.Head
			} else if id == int(*nodeCount)-1 {
				node.Position = types.Tail
			} else {
				node.Position = types.Middle
			}

			leftBuf := simulator.NewRoundBuffer(node.LeftInbox)
			rightBuf := simulator.NewRoundBuffer(node.RightInbox)

			// 1. Setup (Network + Discovery)
			if err := simulator.SetupNode(node); err != nil {
				fmt.Printf("Error setup node %d: %v\n", id, err)
				return
			}

			// Slight delay to ensure listeners are active before discovery flooding
			time.Sleep(100 * time.Millisecond)

			if err := engine.InitialSetup(node, cfg, leftBuf, rightBuf); err != nil {
				fmt.Printf("Error discovery node %d: %v\n", id, err)
				return
			}

			// 2. Generate Value
			val := shared.GenerateInitialValue(id, cfg)
			node.Value = algorithms.OddEvenPayload{Value: val}

			// 3. Run
			algorithms.RunOddEven(node, engine, leftBuf, rightBuf, transport.SendMessage[algorithms.OddEvenPayload])

		}(i)
	}

	wg.Wait()
	fmt.Println("Simulation Complete.")
}
