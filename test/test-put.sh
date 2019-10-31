UNIQID=$1
SIZE=$2

# echo "PUT #$UNIQID: Generating Test Files..."
rm -f test-even-$UNIQID.dat
rm -f test-odd-$UNIQID.dat
dd if=/dev/random of=./test-even-$UNIQID.dat bs=512 count=$SIZE 2> /dev/null
dd if=/dev/random of=./test-odd-$UNIQID.dat bs=511 count=$SIZE 2> /dev/null

# echo "PUT #$UNIQID: PUT to TFTP Server..."
rm -f put-$UNIQID.out
echo "binary" > put-$UNIQID.out
echo "put test-even-$UNIQID.dat" >> put-$UNIQID.out
echo "put test-odd-$UNIQID.dat" >> put-$UNIQID.out
echo "quit" >> put-$UNIQID.out
cat put-$UNIQID.out | tftp 127.0.0.1 > /dev/null

md5sum test-even-$UNIQID.dat > put-md5sum-$UNIQID.out
md5sum test-odd-$UNIQID.dat >> put-md5sum-$UNIQID.out

rm -f test-even-$UNIQID.dat test-odd-$UNIQID.dat put-$UNIQID.out
