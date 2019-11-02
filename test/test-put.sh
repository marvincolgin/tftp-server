UNIQID=$1
SIZE=$2

# echo "PUT #$UNIQID: Generating Test Files..."
rm -f $UNIQID-test-even.dat
rm -f $UNIQID-test-odd.dat
dd if=/dev/random of=./$UNIQID-test-even.dat bs=512 count=$SIZE 2> /dev/null
dd if=/dev/random of=./$UNIQID-test-odd.dat bs=511 count=$SIZE 2> /dev/null

# echo "PUT #$UNIQID: PUT to TFTP Server..."
rm -f $UNIQID-put.out
echo "binary" > $UNIQID-put.out
echo "put $UNIQID-test-even.dat" >> $UNIQID-put.out
echo "put $UNIQID-test-odd.dat" >> $UNIQID-put.out
echo "quit" >> $UNIQID-put.out
cat $UNIQID-put.out | tftp 127.0.0.1 > /dev/null

md5sum $UNIQID-test-even.dat > $UNIQID-put-md5sum.out
md5sum $UNIQID-test-odd.dat >> $UNIQID-put-md5sum.out

rm -f $UNIQID-test-even.dat $UNIQID-test-odd.dat $UNIQID-put.out
