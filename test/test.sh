# This script will fire off lots of test clients

if [ -z "$1" ]
    then
        echo Param1 must be Number-of-Clients
        exit 1
fi
NUMCLIENTS=$1
echo $NUMCLIENTS

for (( c=1; c<=$NUMCLIENTS; c++ ))
do
    echo "Spawning $c"
    ./test-entrypoint.sh $c &

done
