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
	"github.com/InventedSarawak/Distributed-Sorting-Sim/pkg/types"
)

func main() {
	nodeCount := flag.Uint("node-count", 10, "Number of Nodes")
	inputTypeStr := flag.String("input-type", "random", "Input type")
	debug := flag.Bool("debug", false, "Enable verbose logging")
	benchmark := flag.Bool("benchmark", false, "Enable benchmarking metrics")
	flag.Parse()

	if *debug {
		fmt.Printf("--- Distributed Sorting Simulator (Odd-Even) ---\n")
		fmt.Printf("Nodes: %d | Type: %s\n", *nodeCount, *inputTypeStr)
	}

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
	var startTime time.Time

	initialArray := make([]int, *nodeCount)
	finalArray := make([]int, *nodeCount)

	if *benchmark {
		startTime = time.Now()
	}

	for i := 0; i < int(*nodeCount); i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			node := &types.Node[algorithms.OddEvenPayload]{
				ID:         id,
				TotalNode:  int(*nodeCount),
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

			if err := simulator.SetupNode(node, *debug); err != nil {
				if *debug {
					fmt.Printf("Error setup node %d: %v\n", id, err)
				}
				return
			}

			time.Sleep(100 * time.Millisecond)

			if err := engine.InitialSetup(node, cfg, leftBuf, rightBuf, *debug); err != nil {
				if *debug {
					fmt.Printf("Error discovery node %d: %v\n", id, err)
				}
				return
			}

			val := simulator.GenerateInitialValue(id, cfg)
			node.Value = algorithms.OddEvenPayload{Value: val}

			initialArray[id] = val

			algorithms.RunOddEven(node, engine, leftBuf, rightBuf, transport.SendMessage[algorithms.OddEvenPayload], *debug)

			finalArray[id] = node.Value.Value
		}(i)
	}

	wg.Wait()

	if *nodeCount <= 100 {
		fmt.Printf("\n--- Results (N=%d) ---\n", *nodeCount)
		fmt.Printf("Initial: %v\n", initialArray)
		fmt.Printf("Final:   %v\n", finalArray)
	}

	if *benchmark {
		elapsed := time.Since(startTime)
		fmt.Printf("\n--- Benchmark Results ---\n")
		fmt.Printf("Algorithm: Odd-Even Transposition\n")
		fmt.Printf("Nodes: %d\n", *nodeCount)
		fmt.Printf("Time: %v\n", elapsed)
	} else {
		fmt.Println("Simulation Complete.")
	}
}
