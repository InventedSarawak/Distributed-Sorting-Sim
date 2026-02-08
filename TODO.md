# 🚀 Distributed Sorting Simulator: Execution Checklist

## 1. Project Initialization & Setup
- [ ] Initialize Go module: `go mod init <your-repo-name>`
- [ ] Create folder structure:
    - `cmd/`, `internal/`, `pkg/`, `bin/`, `scripts/`, `outputs/`
- [ ] Define shared message types in `pkg/types/message.go` (Value, Step, Type).
- [ ] Implement TCP Node primitives in `internal/transport/tcp_node.go`:
    - [ ] `Listen()` logic for incoming neighbor connections.
    - [ ] `Connect()` logic to link to left/right neighbors.
    - [ ] `Send/Receive` helpers with JSON encoding.
    - [ ] Barrier Synchronization logic (to keep rounds in sync).

---

## 2. Algorithm Implementation (`internal/algorithms/`)
### (a) Odd-Even Transposition
- [ ] Implement `OddEvenStep()`:
    - [ ] Even Round: Nodes (2i, 2i+1) exchange and compare.
    - [ ] Odd Round: Nodes (2i+1, 2i+2) exchange and compare.
- [ ] Verify sorting for $n=10$ locally.

### (b) Sasaki's Time-Optimal Algorithm
- [ ] Implement State Machine logic (handling the $n$ time-optimal rounds).
- [ ] Handle simultaneous bidirectional value movement.
- [ ] Verify sorting for $n=10$ locally.

### (c) Alternative: Pipelined Min-Max (or Prasath's)
- [ ] Implement logic for the chosen alternative.
- [ ] Ensure it completes in $O(n)$ rounds.
- [ ] Verify sorting for $n=10$ locally.

---

## 3. The Orchestrator (`cmd/`)
- [ ] Create `main.go` for each algorithm in `cmd/`:
    - [ ] Support CLI flags for `-n` (number of nodes) and `-port-start`.
    - [ ] Implement "Spawner" logic to trigger $n$ nodes.
    - [ ] Implement "Result Collector" to gather final values and verify sort order.

---

## 4. Benchmarking & Data Collection
- [ ] Write `scripts/run_benchmarks.sh` to automate:
    - [ ] $n = 1000$
    - [ ] $n = 2000$
    - [ ] $n = 3000$
    - [ ] $n = 5000$
- [ ] Capture "Wall Clock Time" for each run.
- [ ] Save output logs for all 3 algorithms to `outputs/`.

---

## 5. Documentation & Final Polish
- [ ] **README.txt**:
    - [ ] Add compilation instructions (`go build ...`).
    - [ ] Add run instructions for each algorithm.
- [ ] **Report**:
    - [ ] Create the **Comparison Table** ($n$ vs Time).
    - [ ] Document **Time Complexity** ($O(n)$).
    - [ ] Document **Space Complexity** ($O(1)$ per node + choice of data structures).
- [ ] **Validation**:
    - [ ] Check that all file names are lowercase.
    - [ ] Ensure TCP ports are released correctly after each run.

---

## 6. Submission Packaging
- [ ] Compile all 3 programs and verify they are in the zip.
- [ ] Verify Roll No & Year (e.g., `S2024...`).
- [ ] Name the file: `YYYY-abc-dc-assign01.zip` (Check all lowercase).
- [ ] **Final Check**: Does the zip contain:
    - [ ] 3 Programs?
    - [ ] README.txt?
    - [ ] Results/Output files?
    - [ ] Report with complexities?