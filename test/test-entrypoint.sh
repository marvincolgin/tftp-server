#/bin/bash

if [ -z "$1" ]
    then
        echo Param1 must be unique-id
        exit 1
fi

UNIQID=$1
SIZE=$2
SIZE=${SIZE:="10000"}

# echo "Testing..."
./test-put.sh $UNIQID $SIZE
./test-get.sh $UNIQID


# echo "Comparison..."
diff put-md5sum-$UNIQID.out get-md5sum-$UNIQID.out > final-$UNIQID.out
filesize=$(wc -c "final-$UNIQID.out" | awk '{print $1}')
if [ $filesize -ne 0 ]; then
    echo "ERROR #$UNIQID: MISMATCH MD5SUM!!!"
    cat final-$UNIQID.out
else
    echo "OK #$UNIQID: Perfect Match"
    rm final-$UNIQID.out
fi

