#!/bin/bash

if [ ! -f "bin/odd_even" ] || [ ! -f "bin/sasaki" ] || [ ! -f "bin/alternative" ]; then
    echo "Error: Binaries not found. Please run 'task build' or 'go build' first."
    exit 1
fi

NODE_COUNTS=(10 50 100 500)
INPUT_TYPE="random"

echo "=========================================================="
echo "   Starting Distributed Sorting Benchmark"
echo "   Input Type: $INPUT_TYPE"
echo "=========================================================="

for N in "${NODE_COUNTS[@]}"; do
    echo ""
    echo "----------------------------------------------------------"
    echo "    Testing Node Count: $N"
    echo "----------------------------------------------------------"

    echo -n "Running Odd-Even... "
    ./bin/odd_even -node-count $N -input-type $INPUT_TYPE -benchmark | grep "Time:" || echo "Failed"

    echo -n "Running Sasaki...   "
    ./bin/sasaki -node-count $N -input-type $INPUT_TYPE -benchmark | grep "Time:" || echo "Failed"

    echo -n "Running Alternative..."
    ./bin/alternative -node-count $N -input-type $INPUT_TYPE -benchmark | grep "Time:" || echo "Failed"
done

echo ""
echo "=========================================================="
echo "    Benchmark Complete"
echo "=========================================================="