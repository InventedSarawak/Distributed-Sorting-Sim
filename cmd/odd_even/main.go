package main

import (
	"flag"
	"fmt"
	"strings"
	"sync"

	"github.com/InventedSarawak/Distributed-Sorting-Sim/internal/algorithms"
	"github.com/InventedSarawak/Distributed-Sorting-Sim/internal/simulator" // Fixed: Import the simulator package
	"github.com/InventedSarawak/Distributed-Sorting-Sim/pkg/shared"
	"github.com/InventedSarawak/Distributed-Sorting-Sim/pkg/types"
)

func main() {
	// Define flags for node count and input type
	nodeCount := flag.Uint("node-count", 50, "Enter the number of Nodes")
	inputTypeStr := flag.String("input-type", "random", "Enter the type of input (random, sorted, reverse)")
	flag.Parse()

	// Validate node count
	if *nodeCount <= 0 || *nodeCount > 7000 {
		flag.Usage()
		panic("Invalid node count. Must be between 1 and 7000.")
	}

	// Map string flag to InputType constant
	var inputType types.InputType
	switch strings.ToLower(*inputTypeStr) {
	case "random":
		inputType = types.Random
	case "sorted":
		inputType = types.Sorted
	case "reverse":
		inputType = types.Reverse
	default:
		flag.Usage()
		panic("Invalid input type. Must be one of: random, sorted, reverse.")
	}

	// Initialize configuration
	cfg := types.Config{
		NodeCount: uint(*nodeCount),
		InputType: inputType, // Fixed: Using the inputType variable resolves the UnusedVar error
		Algorithm: types.OddEven,
	}

	// Create the simulation engine
	engine := simulator.NewEngine[algorithms.OddEvenPayload](int(*nodeCount))
	var wg sync.WaitGroup

	// Spawn N nodes
	for i := 0; i < int(*nodeCount); i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Initialize Node Struct
			node := &types.Node[algorithms.OddEvenPayload]{
				ID:         id,
				TotalNode:  int(*nodeCount),
				LeftInbox:  make(chan types.Message[algorithms.OddEvenPayload], 100),
				RightInbox: make(chan types.Message[algorithms.OddEvenPayload], 100),
			}

			// Assign Position based on ID
			if id == 0 {
				node.Position = types.Head
			} else if id == int(*nodeCount)-1 {
				node.Position = types.Tail
			} else {
				node.Position = types.Middle
			}

			// 1. Setup Persistent Connections & Dispatchers
			// Ensure these functions are defined in your internal/simulator package
			leftBuf := simulator.NewRoundBuffer(node.LeftInbox)
			rightBuf := simulator.NewRoundBuffer(node.RightInbox)

			if err := simulator.SetupNode(node); err != nil {
				fmt.Printf("Error setting up node %d: %v\n", id, err)
				return
			}

			// 2. Initial Setup (Discovery & Generation)
			// The generator uses the config.InputType to decide how to fill node.Value
			shared.GenerateInitialValue(node, cfg)

			// 3. Run Algorithm
			algorithms.RunOddEven(node, engine, leftBuf, rightBuf)

			fmt.Printf("Node %d finished. Final Value: %v\n", id, node.Value.Value)
		}(i)
	}

	wg.Wait()
	fmt.Println("Simulation Complete.")
}
