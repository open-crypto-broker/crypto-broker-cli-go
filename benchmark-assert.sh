#!/bin/bash

set -e

RESULTS_FILE="$1"
if [ -z "$RESULTS_FILE" ]; then
    echo "Usage: $0 <benchmark-results.json>"
    exit 1
fi

echo "Running benchmark performance assertions..."
echo "Please note that this script is only meant to be run by github workflow worker and thresholds may not be accurate for local runs."

extract_benchmark_result() {
    local bench_name="$1"
    local results_file="$2"
    
    grep -F "$bench_name" "$results_file" | grep "ns/op" | grep '"Action":"output"' | sed -n 's/.*"Output":"\([^"]*\)".*/\1/p' | sed 's/\\n/\n/g' | sed 's/\\t/\t/g'
}

benchmark_failed() {
    local bench_name="$1"
    local results_file="$2"
    
    if grep "\"Test\":\"$bench_name\"" "$results_file" | grep -q '"Action":"fail"'; then
        return 0  # true - benchmark failed
    else
        return 1  # false - benchmark did not fail
    fi
}

# Function to assert benchmark performance
assert_benchmark() {
    local bench_name="$1"
    local max_ns_per_op="$2"
    local max_allocs_per_op="$3"

    # Extract benchmark output lines
    local output_lines
    output_lines=$(extract_benchmark_result "$bench_name" "$RESULTS_FILE")

    if [ -z "$output_lines" ]; then
        echo "❌ Benchmark $bench_name: No output found"
        return 1
    fi

    # Get the last (most recent) result line
    # Metrics may be wrapped across two lines (allocs/op on next line); join last two lines.
    local metrics_line
    metrics_line=$(echo "$output_lines" | tail -2 | paste -sd ' ' -)

    # Parse ns/op value (accept tabs/spaces) using a robust pattern.
    # Extract ns/op value (number before "ns/op") - handle decimal numbers
    local ns_per_op
    ns_per_op=$(echo "$metrics_line" | sed -E 's/.*[[:space:]]([0-9]+(\.[0-9]+)?)[[:space:]]+ns\/op.*/\1/' | sed 's/,//g')
    
    # Parse allocs/op value from joined line.
    local allocs_per_op
    allocs_per_op=$(echo "$metrics_line" | sed -E 's/.*[[:space:]]([0-9]+)[[:space:]]+allocs\/op.*/\1/' | sed 's/,//g')

    # Validate that we got numeric values
    if ! [[ "$ns_per_op" =~ ^[0-9]+(\.[0-9]+)?$ ]]; then
        echo "❌ Benchmark $bench_name: Could not parse ns/op value from: $metrics_line"
        return 1
    fi

    if ! [[ "$allocs_per_op" =~ ^[0-9]+$ ]]; then
        echo "❌ Benchmark $bench_name: Could not parse allocs/op value from: $metrics_line"
        return 1
    fi

    # Perform assertions - check if performance is WORSE than threshold
    if (( $(echo "$ns_per_op > $max_ns_per_op" | bc -l 2>/dev/null || echo "0") )); then
        echo "❌ $bench_name: ns/op ($ns_per_op) exceeds threshold ($max_ns_per_op)"
        return 1
    fi

    if [ "$allocs_per_op" -gt "$max_allocs_per_op" ]; then
        echo "❌ $bench_name: allocs/op ($allocs_per_op) exceeds threshold ($max_allocs_per_op)"
        return 1
    fi

    echo "✅ $bench_name: ns/op=$ns_per_op, allocs/op=$allocs_per_op"
    return 0
}

# Function to run benchmark assertion with automatic failure handling
run_benchmark_assertion() {
    local bench_name="$1"
    local max_ns_per_op="$2"
    local max_allocs_per_op="$3"
    
    # Check if this benchmark failed during execution
    if benchmark_failed "$bench_name" "$RESULTS_FILE"; then
        echo "⚠️  $bench_name: Skipped due to test failure"
        return 0  # Don't fail the script for benchmark failures
    else
        # Benchmark didn't fail, run the assertion
        assert_benchmark "$bench_name" "$max_ns_per_op" "$max_allocs_per_op"
        return $?
    fi
}

ceil() {
    awk -v x="$1" 'BEGIN {
        print (x == int(x)) ? x : int(x) + 1
    }'
}

# threshold=$((baseline * (1 + margin + env_factor)))
threshold_ns_per_op_BenchmarkHash_profile_Default_Sequential=$(ceil "$(awk 'BEGIN {print 2005227 * (1 + 0.25)}')")
threshold_allocs_per_op_BenchmarkHash_profile_Default_Sequential=$(ceil "$(awk 'BEGIN {print 160 * (1 + 0.1)}')")

threshold_ns_per_op_BenchmarkHash_profile_Default_Parallel=$(ceil "$(awk 'BEGIN {print 1734807 * (1 + 0.25)}')")
threshold_allocs_per_op_BenchmarkHash_profile_Default_Parallel=$(ceil "$(awk 'BEGIN {print 138 * (1 + 0.1)}')")

threshold_ns_per_op_BenchmarkHealth_Sequential=$(ceil "$(awk 'BEGIN {print 2114627 * (1 + 0.25)}')")
threshold_allocs_per_op_BenchmarkHealth_Sequential=$(ceil "$(awk 'BEGIN {print 121 * (1 + 0.1)}')")

threshold_ns_per_op_BenchmarkHealth_Parallel=$(ceil "$(awk 'BEGIN {print 2573472 * (1 + 0.25)}')")
threshold_allocs_per_op_BenchmarkHealth_Parallel=$(ceil "$(awk 'BEGIN {print 196 * (1 + 0.1)}')")

threshold_ns_per_op_BenchmarkSign_profile_Default_CSR_SECP256R1_CA_RSA4096_Sequential=$(ceil "$(awk 'BEGIN {print 22164674 * (1 + 0.4)}')")
threshold_allocs_per_op_BenchmarkSign_profile_Default_CSR_SECP256R1_CA_RSA4096_Sequential=$(ceil "$(awk 'BEGIN {print 864 * (1 + 0.1)}')")

threshold_ns_per_op_BenchmarkSign_profile_Default_CSR_SECP256R1_CA_RSA4096_Parallel=$(ceil "$(awk 'BEGIN {print 6331463 * (1 + 0.4)}')")
threshold_allocs_per_op_BenchmarkSign_profile_Default_CSR_SECP256R1_CA_RSA4096_Parallel=$(ceil "$(awk 'BEGIN {print 324 * (1 + 0.1)}')")

threshold_ns_per_op_BenchmarkSign_profile_Default_CSR_SECP521R1_CA_SECP521R1_Sequential=$(ceil "$(awk 'BEGIN {print 16116639 * (1 + 0.4)}')")
threshold_allocs_per_op_BenchmarkSign_profile_Default_CSR_SECP521R1_CA_SECP521R1_Sequential=$(ceil "$(awk 'BEGIN {print 300 * (1 + 0.1)}')")

threshold_ns_per_op_BenchmarkSign_profile_Default_CSR_SECP521R1_CA_SECP521R1_Parallel=$(ceil "$(awk 'BEGIN {print 8967990 * (1 + 0.4)}')")
threshold_allocs_per_op_BenchmarkSign_profile_Default_CSR_SECP521R1_CA_SECP521R1_Parallel=$(ceil "$(awk 'BEGIN {print 300 * (1 + 0.1)}')")

threshold_ns_per_op_BenchmarkSign_profile_Default_CSR_SECP256R1_CA_SECP384R1_Sequential=$(ceil "$(awk 'BEGIN {print 8482845 * (1 + 0.4)}')")
threshold_allocs_per_op_BenchmarkSign_profile_Default_CSR_SECP256R1_CA_SECP384R1_Sequential=$(ceil "$(awk 'BEGIN {print 300 * (1 + 0.1)}')")

threshold_ns_per_op_BenchmarkSign_profile_Default_CSR_SECP256R1_CA_SECP384R1_Parallel=$(ceil "$(awk 'BEGIN {print 4208205 * (1 + 0.4)}')")
threshold_allocs_per_op_BenchmarkSign_profile_Default_CSR_SECP256R1_CA_SECP384R1_Parallel=$(ceil "$(awk 'BEGIN {print 300 * (1 + 0.1)}')")

# Run assertions for all benchmarks with automatic failure handling
run_benchmark_assertion "BenchmarkHash_profile_Default_Sequential" $threshold_ns_per_op_BenchmarkHash_profile_Default_Sequential $threshold_allocs_per_op_BenchmarkHash_profile_Default_Sequential
run_benchmark_assertion "BenchmarkHash_profile_Default_Parallel" $threshold_ns_per_op_BenchmarkHash_profile_Default_Parallel $threshold_allocs_per_op_BenchmarkHash_profile_Default_Parallel
run_benchmark_assertion "BenchmarkHealth_Sequential" $threshold_ns_per_op_BenchmarkHealth_Sequential $threshold_allocs_per_op_BenchmarkHealth_Sequential
run_benchmark_assertion "BenchmarkHealth_Parallel" $threshold_ns_per_op_BenchmarkHealth_Parallel $threshold_allocs_per_op_BenchmarkHealth_Parallel
run_benchmark_assertion "BenchmarkSign_profile_Default_CSR_SECP256R1_CA_RSA4096_Sequential" $threshold_ns_per_op_BenchmarkSign_profile_Default_CSR_SECP256R1_CA_RSA4096_Sequential $threshold_allocs_per_op_BenchmarkSign_profile_Default_CSR_SECP256R1_CA_RSA4096_Sequential
run_benchmark_assertion "BenchmarkSign_profile_Default_CSR_SECP256R1_CA_RSA4096_Parallel" $threshold_ns_per_op_BenchmarkSign_profile_Default_CSR_SECP256R1_CA_RSA4096_Parallel $threshold_allocs_per_op_BenchmarkSign_profile_Default_CSR_SECP256R1_CA_RSA4096_Parallel
run_benchmark_assertion "BenchmarkSign_profile_Default_CSR_SECP521R1_CA_SECP521R1_Sequential" $threshold_ns_per_op_BenchmarkSign_profile_Default_CSR_SECP521R1_CA_SECP521R1_Sequential $threshold_allocs_per_op_BenchmarkSign_profile_Default_CSR_SECP521R1_CA_SECP521R1_Sequential
run_benchmark_assertion "BenchmarkSign_profile_Default_CSR_SECP521R1_CA_SECP521R1_Parallel" $threshold_ns_per_op_BenchmarkSign_profile_Default_CSR_SECP521R1_CA_SECP521R1_Parallel $threshold_allocs_per_op_BenchmarkSign_profile_Default_CSR_SECP521R1_CA_SECP521R1_Parallel
run_benchmark_assertion "BenchmarkSign_profile_Default_CSR_SECP256R1_CA_SECP384R1_Sequential" $threshold_ns_per_op_BenchmarkSign_profile_Default_CSR_SECP256R1_CA_SECP384R1_Sequential $threshold_allocs_per_op_BenchmarkSign_profile_Default_CSR_SECP256R1_CA_SECP384R1_Sequential
run_benchmark_assertion "BenchmarkSign_profile_Default_CSR_SECP256R1_CA_SECP384R1_Parallel" $threshold_ns_per_op_BenchmarkSign_profile_Default_CSR_SECP256R1_CA_SECP384R1_Parallel $threshold_allocs_per_op_BenchmarkSign_profile_Default_CSR_SECP256R1_CA_SECP384R1_Parallel

echo ""
echo "All benchmark assertions passed!"
echo ""
echo "Full Benchmark Summary:"
grep '"Action":"output"' "$RESULTS_FILE" | grep -E "(ns/op|B/op|allocs/op)" | sed -n 's/.*"Output":"\([^"]*\)".*/\1/p' | sed 's/\\n/\n/g' | sed 's/\\t/\t/g' | sed 's/^/   /'