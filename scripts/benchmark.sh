#!/bin/bash

if [ ! -f "bin/odd_even" ] || [ ! -f "bin/sasaki" ] || [ ! -f "bin/alternative" ]; then
    echo "Error: Binaries not found. Please run 'task build' or 'go build' first."
    exit 1
fi

NODE_COUNTS=(1000 2000 3000 5000)
INPUT_TYPE="random"
RESULTS_FILE="results.txt"

# Clear the results file
> "$RESULTS_FILE"

echo "==========================================================" >> "$RESULTS_FILE"
echo "   Starting Distributed Sorting Benchmark" >> "$RESULTS_FILE"
echo "   Input Type: $INPUT_TYPE" >> "$RESULTS_FILE"
echo "==========================================================" >> "$RESULTS_FILE"

for N in "${NODE_COUNTS[@]}"; do
    echo "" >> "$RESULTS_FILE"
    echo "----------------------------------------------------------" >> "$RESULTS_FILE"
    echo "    Testing Node Count: $N" >> "$RESULTS_FILE"
    echo "----------------------------------------------------------" >> "$RESULTS_FILE"

    echo "Running Odd-Even for Node Count $N..."
    ./bin/odd_even -node-count $N -input-type $INPUT_TYPE -benchmark >> "$RESULTS_FILE"

    echo "Running Sasaki for Node Count $N..."
    ./bin/sasaki -node-count $N -input-type $INPUT_TYPE -benchmark >> "$RESULTS_FILE"

    echo "Running Alternative for Node Count $N..."
    ./bin/alternative -node-count $N -input-type $INPUT_TYPE -benchmark >> "$RESULTS_FILE"
done

echo "" >> "$RESULTS_FILE"
echo "==========================================================" >> "$RESULTS_FILE"
echo "    Benchmark Complete" >> "$RESULTS_FILE"
echo "==========================================================" >> "$RESULTS_FILE"