# Distributed Sorting Simulator

This project implements a discrete event simulator for three distributed sorting algorithms on a line network using Go. It simulates a distributed computing environment where processing entities (nodes) communicate via TCP sockets to sort a distributed array of numbers.

## Algorithms Implemented

1.  **Odd-Even Transposition Sort**: A standard parallel sorting algorithm requiring N rounds.
2.  **Sasaki's Time-Optimal Algorithm**: A theoretically optimal algorithm requiring exactly N-1 rounds.
3.  **Alternative Time-Optimal Algorithm (Pipelined Min-Max)**: A pipelined approach using local sorts on triplets.

## System Architecture

The simulation mimics a real distributed network using the following components:

- **Node Representation**: Each processing entity is implemented as a concurrent Go routine (`go func`). Nodes operate asynchronously but synchronize logically via a Global Clock (round number).
- **Communication Layer**:
    - **Transport**: Nodes communicate via TCP sockets (`net.Conn`). Each node binds a listener to a specific port and establishes persistent connections with its immediate neighbors.
    - **Serialization**: Messages are serialized using JSON (`encoding/json`) to ensure a standardized payload format.
    - **Buffering**: A `RoundBuffer` (Hash Map protected by Mutex) handles asynchronous message arrival, storing future round messages until the node is ready.
- **Synchronization**:
    - Nodes maintain a logical clock (`round`).
    - Barrier Synchronization is used within algorithm loops; nodes block until they receive necessary messages from neighbors for the current round.

## Project Structure

```.
├── cmd
│   ├── odd_even
│   │   └── main.go
│   ├── sasaki
│   │   └── main.go
│   └── alternative
│       └── main.go
├── internal
│   ├── algorithms
│   │   ├── odd_even.go
│   │   ├── sasaki.go
│   │   └── alternative.go
│   ├── simulator
|   |   ├── barrier.go
│   │   └── engine.go
│   └── transport
|       ├── tcp_node.go
│       └── tcp_node_test.go
├── pkg
│   └── types
│       └── message.go
├── scripts
│   └── benchmark.sh
├── go.mod
├── .gitignore
├── Taskfile.yml
└── README.md
```

## Complexity Analysis & Data Structures

### 1. Odd-Even Transposition Algorithm

- **Time Complexity**: O(N) rounds.
- **Space Complexity**: O(1) per process.
- **Data Structures**: Simple payload containing a single `int` value. Uses two buffered channels (`LeftInbox`, `RightInbox`) as queues.
- **Implementation**: Even-indexed nodes exchange with right neighbors in even rounds; odd-indexed nodes exchange in odd rounds. This implementation has the lowest CPU overhead due to minimal payload size.

### 2. Sasaki's Time-Optimal Algorithm

- **Time Complexity**: O(N) (Theoretically N-1 rounds).
- **Space Complexity**: O(1) per process.
- **Data Structures**: Complex payload `SasakiPayload` containing:
    - `LValue` (struct: Value int, IsMarked bool)
    - `RValue` (struct: Value int, IsMarked bool)
    - `Area` (int) - Counter tracking sorted segments.
- **Implementation**: Involves a 3-phase process per round (Probe, Barrier, Update). The complex state updates and larger JSON payloads result in higher CPU and I/O costs compared to Odd-Even in this simulation.

### 3. Alternative Time-Optimal Algorithm (Pipelined Min-Max)

- **Time Complexity**: O(N).
- **Space Complexity**: O(1) per process.
- **Data Structures**: Simple `int` payload. Center nodes use a temporary local buffer (slice of size 3) to sort triplets.
- **Implementation**: Nodes assume dynamic roles (Center, Left-Wing, Right-Wing) based on `(round + 1) % 3`. Center nodes block waiting for _both_ neighbors, making this algorithm sensitive to scheduler latency.

## Benchmarking Results

The following table summarizes the wall-clock execution time for various node counts (N) on a single-machine simulation.

| N    | Odd-Even Transposition | Sasaki's Time-Optimal | Alternative (Pipelined) |
| ---- | ---------------------- | --------------------- | ----------------------- |
| 10   | 313.17 ms              | 312.67 ms             | 312.62 ms               |
| 50   | 428.64 ms              | 437.68 ms             | 434.50 ms               |
| 100  | 381.63 ms              | 417.50 ms             | 481.13 ms               |
| 500  | 1.32 s                 | 2.43 s                | 1.73 s                  |
| 1000 | 3.96 s                 | 8.86 s                | 5.01 s                  |
| 2000 | 13.76 s                | 32.90 s               | 19.20 s                 |
| 3000 | 30.99 s                | 1m 15.19 s            | 43.95 s                 |
| 5000 | 1m 30.56 s             | 3m 35.15 s            | 2m 1.42 s               |

**Performance Observations:**

- **Small Scale (N < 100)**: Initialization overhead dominates; all algorithms perform similarly (~300-400ms).
- **Large Scale (N > 1000)**: Odd-Even consistently outperforms the others.
    - **Vs. Sasaki**: Sasaki's complex payload results in heavy JSON serialization overhead, outweighing its theoretical round efficiency on a single CPU.
    - **Vs. Alternative**: The Alternative algorithm suffers from pipeline stalls due to strict 3-node synchronization barriers, whereas Odd-Even uses simpler pairwise synchronization.

## Setup and Execution

### Prerequisites

- Go 1.25 or higher installed.

### Compilation

You can compile the binaries manually using the `go build` command:

```bash
mkdir -p bin
go build -o bin/odd_even cmd/odd_even/main.go
go build -o bin/sasaki cmd/sasaki/main.go
go build -o bin/alternative cmd/alternative/main.go
```

Alternatively, if you have the provided scripts:

```bash
chmod +x scripts/benchmark.sh
./scripts/benchmark.sh
```

### Execution

Once compiled, you can run any algorithm using the generated binaries.

#### Syntax:

```bash
./bin/<algorithm_name> -node-count <N> -input-type <type> [flags]
```

#### Examples:

- **Run Odd-Even Sort (Standard):**

```bash
./bin/odd_even -node-count 100 -input-type random
```

- **Run Sasaki's Algorithm (Sorted Input):**

```bash
./bin/sasaki -node-count 100 -input-type sorted
```

- **Run Alternative Algorithm (Debug Mode):**
  _(Recommended only for small N < 20)_

```bash
./bin/alternative -node-count 10 -input-type random -debug
```

#### Flags:

- `-node-count <N>`: Number of nodes in the simulation (required).
- `-input-type <type>`: Type of initial array (random, sorted, reversed) (required).
- `-debug`: Enable verbose logging for debugging purposes (optional).
- `benchmark`: Run the benchmark suite for all algorithms and node counts (optional).

#### Note on High Concurrency

Simulating a large number of nodes (e.g., N=5000) creates thousands of concurrent goroutines and file descriptors. If the simulation hangs or fails, increase your system's file descriptor limit:

```bash
ulimit -n 10000
```
