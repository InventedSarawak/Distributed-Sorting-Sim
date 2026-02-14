# 🚀 Distributed Sorting Simulator: Execution Checklist

## 1. Project Initialization & Setup
- [x] Initialize Go module: `go mod init <your-repo-name>`
- [x] Create folder structure: `cmd/`, `internal/`, `pkg/`, `bin/`, `scripts/`, `outputs/`
- [x] Define shared message types in `pkg/types/message.go`
- [ ] Implement TCP Node primitives in `internal/transport/tcp_node.go`:
    - [x] `Listen()` logic with retry for port binding
    - [x] `SendMessage()` with exponential backoff dialing
    - [x] `HandleConnection()` using `json.NewDecoder` for stream safety
    - [ ] **Persistent Connection Manager**: Refactor `SendMessage` to maintain open `net.Conn` to reduce handshake overhead for $n=5000$.
    - [ ] **Inbox Dispatcher**: Implement background goroutines to route incoming messages to `LeftInbox` or `RightInbox` based on `IncomingDirection`.

---

## 2. Simulator & Synchronization Engine (`internal/simulator/`)
- [ ] **Implement Barrier Synchronization (`barrier.go`)**:
    - [ ] **`RoundBuffer[T]`**: Create a thread-safe mailbox to store "future" messages that arrive before the node is ready for that round.
    - [ ] **`GetStepMessage(targetRound)`**: Implement blocking logic that retrieves a message from the buffer if present, or waits on the inbox channel.
    - [ ] **`WaitForNeighbors(currentRound)`**: Implement a local barrier that ensures a node has received all required neighbor data before proceeding.
- [ ] **Implement Execution Engine (`engine.go`)**:
    - [ ] **Node Lifecycle Manager**: Logic to instantiate `Node[T]` with correct `Position` (Head/Tail/Middle) and ID.
    - [ ] **Logical Clock Management**: Logic to increment `Node.Round` only after all sub-steps of an algorithm phase are verified.
    - [ ] **Topology Handshake**: Implement a pre-start phase where all $n$ nodes verify TCP connectivity before Round 0 begins.

---

## 3. Algorithm Implementation (`internal/algorithms/`)
### (a) Odd-Even Transposition
- [ ] Implement `OddEvenStep()`: Use generic `Message[OddEvenPayload]` to exchange and compare values.
- [ ] Verify sorting for $n=10$ locally.

### (b) Sasaki's Time-Optimal Algorithm
- [ ] Define `SasakiPayload`: Include `Value` and `IsMarked` state.
- [ ] Implement 3-Phase Round logic:
    - [ ] **Phase 1 (Probe)**: Sync `IsMarked` status with neighbors.
    - [ ] **Phase 2 (Data)**: Perform simultaneous bidirectional value movement.
    - [ ] **Phase 3 (Ack)**: Confirm state transitions to prevent race conditions.
- [ ] Verify sorting for $n=10$ locally.

### (c) Alternative: Pipelined Min-Max
- [ ] Implement logic for the chosen alternative in $O(n)$ rounds.
- [ ] Verify sorting for $n=10$ locally.

---

## 4. The Orchestrator (`cmd/`)
- [ ] Create `main.go` for each algorithm: Support `-n` and `-port-start` flags.
- [ ] **Spawner**: Logic to trigger $n$ node goroutines with unique port assignments.
- [ ] **Result Collector**: Logic to aggregate final values from all nodes and verify global sort order.

---

## 5. Benchmarking & Data Collection
- [ ] Write `scripts/run_benchmarks.sh` for $n = 1000, 2000, 3000, 5000$.
- [ ] **Metrics Tracking**:
    - [ ] Capture **Wall Clock Time** per run.
    - [ ] Capture **Total Message Count** (TCP packets) for complexity analysis.
- [ ] Save output logs to `outputs/`.

---

## 6. Documentation & Final Polish
- [ ] **README.txt**: Include `ulimit -n 10000` instructions for high-node simulations.
- [ ] **Report**: Comparison table, time complexity ($O(n)$), and space complexity ($O(1)$ per node).
- [ ] **Validation**: Ensure all files are lowercase and TCP ports release correctly.