#!/bin/bash

# Run the parallel portions of the test
N=(10)
DATA_FOLDER='./data'
OPERATIONS_FOLDER='./object_data/'

for file in `ls ${OPERATIONS_FOLDER} | sort -r`
do
	for n in ${N[@]}
	do
		for i in 1 2 3 4 5
		do
			echo "Running ${file} with ${n} Threads: Iteration ${i}"
			{ (time ./main/sim -i=1000 1000 1000 $n) < $OPERATIONS_FOLDER/$file; } 2>&1 | grep real | cut -f 2 >> ${DATA_FOLDER}/${file}_$n
		done
	done
	echo
done
