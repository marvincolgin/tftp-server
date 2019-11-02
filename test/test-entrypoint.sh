#/bin/bash

if [ -z "$1" ]
    then
        echo Param1 must be unique-id
        exit 1
fi

UNIQID=$1
SIZE=$2
SIZE=${SIZE:="10000"}

ONLYPUT=1

# echo "Testing..."
./test-put.sh $UNIQID $SIZE

if [ $ONLYPUT -ne 1 ]
    then
        ./test-get.sh $UNIQID

        # echo "Comparison..."
        diff $UNIQID-put-md5sum.out $UNIQID-get-md5sum.out > $UNIQID-final.out
        rm $UNIQID-put-md5sum.out $UNIQID-get-md5sum.out
        filesize=$(wc -c "$UNIQID-final.out" | awk '{print $1}')
        if [ $filesize -ne 0 ]; then
            echo "ERROR #$UNIQID: MISMATCH MD5SUM!!!"
            cat $UNIQID-final.out
        else
            echo "OK #$UNIQID: Perfect Match"
            rm $UNIQID-final.out
        fi
    else
        cat $UNIQID-put-md5sum.out
fi
