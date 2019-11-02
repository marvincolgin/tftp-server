# This script will fire off lots of test clients

if [ -z "$1" ]
    then
        echo Param1 must be Number-of-Clients
        exit 1
fi
NUMCLIENTS=$1

SIZE=$2
SIZE=${SIZE:="10000"}

if [ $NUMCLIENTS -eq 1 ]
    then
        ./test-entrypoint.sh $NUMCLIENTS $SIZE
    else
        for c in $(seq -f "%05g" 1 $NUMCLIENTS)
        do
            echo "Spawning $c"
            ./test-entrypoint.sh $c $SIZE &

        done
fi


