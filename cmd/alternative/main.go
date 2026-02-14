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
	nodeCount := flag.Uint("node-count", 10, "Number of Nodes")
	inputTypeStr := flag.String("input-type", "random", "Input type")
	debug := flag.Bool("debug", false, "Enable verbose logging")
	benchmark := flag.Bool("benchmark", false, "Enable benchmarking")
	flag.Parse()

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

	cfg := types.Config{NodeCount: uint(*nodeCount), InputType: inputType, Algorithm: types.Alternate}
	engine := simulator.NewEngine[algorithms.AlternativePayload](int(*nodeCount))
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
			node := &types.Node[algorithms.AlternativePayload]{
				ID: id, TotalNode: int(*nodeCount),
				LeftInbox:  make(chan types.Message[algorithms.AlternativePayload], 500),
				RightInbox: make(chan types.Message[algorithms.AlternativePayload], 500),
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
				return
			}
			time.Sleep(100 * time.Millisecond)
			if err := engine.InitialSetup(node, cfg, leftBuf, rightBuf, *debug); err != nil {
				return
			}

			val := shared.GenerateInitialValue(id, cfg)
			node.Value = algorithms.AlternativePayload{Value: val}

			initialArray[id] = val

			algorithms.RunAlternative(node, engine, leftBuf, rightBuf, transport.SendMessage[algorithms.AlternativePayload], *debug)

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
		fmt.Printf("Alternative Sort Time: %v\n", time.Since(startTime))
	} else {
		fmt.Println("Simulation Complete.")
	}
}
