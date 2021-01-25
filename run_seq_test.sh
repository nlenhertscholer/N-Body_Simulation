#!/bin/bash

# Run the sequential portions of the test
DATA_FOLDER='./data/'
OPERATIONS_FOLDER='./object_data/'

for file in `ls ${OPERATIONS_FOLDER} | sort -r`
do
    for i in 1 2 3 4 5
    do
		echo "Running ${file} Iteration: ${i}"
        { (time ./main/sim -i=1000 1000 1000 0) < $OPERATIONS_FOLDER/${file}; } 2>&1 | grep real | cut -f 2 >> ${DATA_FOLDER}${file}
    done
done
