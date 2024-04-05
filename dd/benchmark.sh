#!/bin/bash

run_benchmark() {
    cmd=$1
    n=$2
    
    total_time=0
    for ((i=1; i<=n; i++)); do

        # clear cache
        if [[ "$clear_cache" == "yes" ]]; then
            sync
            echo 3 > /proc/sys/vm/drop_caches
        fi

        
        output=$(/usr/bin/time -f "%e" sh -c "$cmd" 2>&1 >/dev/null)
    
        time=$(echo "$output" | tail -n1)
        if [[ -z "$time" ]]; then
            echo "Error: Failed to retrieve execution time for command: $cmd" >&2
            return 1
        fi

        total_time=$(echo "$total_time + $time" | bc)
    done
    
    avg_time=$(echo "scale=2; $total_time / $n" | bc)
    echo "$avg_time"
}


original_dd="/usr/bin/dd"
my_dd="./dd"
input_file="testfile.bin"
output_file="output.bin"
block_sizes=("4k" "64k" "1M")
iterations=5
clear_cache="yes"

dd if=/dev/urandom of=$input_file bs=1M count=1024

printf "%-10s %-15s %-15s %-10s\n" "Block Size" "coreutils dd" "great dd" "Speedup"
printf "%-10s %-15s %-15s %-10s\n" "----------" "---------------" "---------------" "----------"

for bs in "${block_sizes[@]}"; do
    original_time=$(run_benchmark "$original_dd if=$input_file of=$output_file bs=$bs" $iterations $clear_cache)
    if [[ $? -ne 0 ]]; then
        echo "Error: Benchmark failed for original dd with block size $bs"
        exit 1
    fi
    
    my_time=$(run_benchmark "$my_dd -if=$input_file -of=$output_file -bs=$bs" $iterations $clear_cache)
    if [[ $? -ne 0 ]]; then
        echo "Error: Benchmark failed for my dd with block size $bs"
        exit 1
    fi
    
    speedup=$(echo "scale=2; $original_time / $my_time" | bc)
    
    printf "%-10s %-15s %-15s %-10.2fx\n" "$bs" "${original_time}s" "${my_time}s" "$speedup"
done

rm $input_file $output_file