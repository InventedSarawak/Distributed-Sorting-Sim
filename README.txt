Assignment-1: Distributed Sorting Simulation
============================================

FILES SUBMITTED:
1. cmd/odd_even/main.go      - Odd-Even Transposition Sort implementation
2. cmd/sasaki/main.go        - Sasaki's Time-Optimal Sort implementation
3. cmd/alternative/main.go   - Alternative (Pipelined) Sort implementation
4. internal/                 - Shared simulation engine, transport, and logic
5. pkg/                      - Shared types and utilities
6. REPORT.txt                - Complexity analysis and results table
7. assignment_outputs/       - Output logs from the execution

HOW TO COMPILE:
----------------
Prerequisites: Go 1.25+ installed.

Option 1 (Using the provided script):
   chmod +x scripts/benchmark.sh
   ./scripts/benchmark.sh
   (This will compile binaries into the bin/ folder and run the benchmark)

Option 2 (Manual Compilation):
   go build -o bin/odd_even cmd/odd_even/main.go
   go build -o bin/sasaki cmd/sasaki/main.go
   go build -o bin/alternative cmd/alternative/main.go

HOW TO RUN:
-----------
To run a specific algorithm (e.g., Odd-Even) for 1000 nodes:

   ./bin/odd_even -node-count 1000 -input-type random -benchmark

Flags:
   -node-count <int> : Number of processing entities (default 10)
   -input-type <str> : 'random', 'sorted', or 'reverse'
   -benchmark        : Runs in silent mode and prints wall-clock time
   -debug            : Prints step-by-step logs (Recommended only for N < 20)

NOTE ON LARGE N (N=5000):
-------------------------
Running 5000 nodes creates 5000 concurrent goroutines and multiple file descriptors per node.
You may need to increase your system limit before running:
   ulimit -n 10000